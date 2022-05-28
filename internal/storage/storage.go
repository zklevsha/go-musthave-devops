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

func (s *MemoryStorage) SetGauge(metricName string, metricValue float64) error {
	s.gaugesMx.Lock()
	s.gauges[metricName] = metricValue
	s.gaugesMx.Unlock()
	return nil
}

func (s *MemoryStorage) GetAllGauges() (map[string]float64, error) {
	c := make(map[string]float64)
	s.gaugesMx.RLock()
	for k, v := range s.gauges {
		c[k] = v
	}
	s.gaugesMx.RUnlock()
	return c, nil
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

func (s *MemoryStorage) IncreaseCounter(metricName string, metricValue int64) error {
	if _, ok := s.counters[metricName]; ok {
		s.countersMx.Lock()
		s.counters[metricName] += metricValue
		s.countersMx.Unlock()
	} else {
		s.countersMx.Lock()
		s.counters[metricName] = metricValue
		s.countersMx.Unlock()
	}
	return nil
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

func (s *MemoryStorage) SetCounter(metricName string, metricValue int64) error {
	s.countersMx.Lock()
	s.counters[metricName] = metricValue
	s.countersMx.Unlock()
	return nil
}

func (s *MemoryStorage) GetAllCounters() (map[string]int64, error) {
	c := make(map[string]int64)
	s.countersMx.RLock()
	for k, v := range s.counters {
		c[k] = v
	}
	s.countersMx.RUnlock()
	return c, nil
}

func (s *MemoryStorage) Avaliable() error {
	return nil

}

func (s *MemoryStorage) Close() {
}

func (s *MemoryStorage) Init() error {
	return nil
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
	SetGauge(metricName string, metricValue float64) error
	GetAllGauges() (map[string]float64, error)
	GetCounter(metricName string) (int64, error)
	SetCounter(metricName string, metricValue int64) error
	IncreaseCounter(metricName string, metricValue int64) error
	GetAllCounters() (map[string]int64, error)
	ResetCounter(metricName string) error
	Avaliable() error
	Close()
	Init() error
}

var Agent = NewMemoryStorage()
