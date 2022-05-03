package storage

import (
	"fmt"
	"sync"
)

type MemoryStorage struct {
	counters   map[string]int64
	gauges     map[string]float64
	countersMx sync.RWMutex
	gaugesMx   sync.RWMutex
}

func (s *MemoryStorage) GetGauge(metricName string) (float64, error) {
	s.gaugesMx.RLock()
	v, ok := s.gauges[metricName]
	s.gaugesMx.RUnlock()
	if ok {
		return v, nil
	} else {
		e := fmt.Errorf("counter metric %s does not exists", metricName)
		return -1, e
	}
}

func (s *MemoryStorage) SetGauge(metricName string, metricValue float64) {
	s.gaugesMx.Lock()
	s.gauges[metricName] = metricValue
	s.gaugesMx.Unlock()
}

func (s *MemoryStorage) GetAllGauges() map[string]float64 {
	c := make(map[string]float64)
	s.gaugesMx.RLock()
	for k, v := range s.gauges {
		c[k] = v
	}
	s.gaugesMx.RUnlock()
	return c
}

func (s *MemoryStorage) GetCounter(metricName string) (int64, error) {
	s.countersMx.RLock()
	v, ok := s.counters[metricName]
	s.countersMx.RUnlock()
	if ok {
		return v, nil
	} else {
		e := fmt.Errorf("counter metric %s does not exists", metricName)
		return -1, e
	}
}

func (s *MemoryStorage) IncreaseCounter(metricName string, metricValue int64) {
	if _, ok := s.counters[metricName]; ok {
		s.countersMx.Lock()
		s.counters[metricName] += metricValue
		s.countersMx.Unlock()
	} else {
		s.countersMx.Lock()
		s.counters[metricName] = metricValue
		s.countersMx.Unlock()
	}
}

func (s *MemoryStorage) ResetCounter(metricName string) error {
	if _, ok := s.counters[metricName]; ok {
		s.countersMx.Lock()
		s.counters[metricName] = 0
		s.countersMx.Unlock()
		return nil
	} else {
		return fmt.Errorf("counter metric %s does not exists", metricName)
	}

}

func (s *MemoryStorage) SetCounter(metricName string, metricValue int64) {
	s.countersMx.Lock()
	s.counters[metricName] = metricValue
	s.countersMx.Unlock()
}

func (s *MemoryStorage) GetAllCounters() map[string]int64 {
	c := make(map[string]int64)
	s.countersMx.RLock()
	for k, v := range s.counters {
		c[k] = v
	}
	s.countersMx.RUnlock()
	return c
}

func NewMemoryStorage() Storage {
	return &MemoryStorage{
		counters:   map[string]int64{},
		gauges:     map[string]float64{},
		gaugesMx:   sync.RWMutex{},
		countersMx: sync.RWMutex{},
	}

}

type Storage interface {
	GetGauge(metricName string) (float64, error)
	SetGauge(metricName string, metricValue float64)
	GetAllGauges() map[string]float64
	GetCounter(metricName string) (int64, error)
	SetCounter(metricName string, metricValue int64)
	IncreaseCounter(metricName string, metricValue int64)
	GetAllCounters() map[string]int64
	ResetCounter(metricName string) error
}

var Server = NewMemoryStorage()
var Agent = NewMemoryStorage()
