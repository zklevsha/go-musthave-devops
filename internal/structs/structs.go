package structs

import (
	"fmt"
	"log"
	"sync"

	"github.com/zklevsha/go-musthave-devops/internal/hash"
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // hmac метрики
}

func (m Metric) CalculateHash(key string) string {
	var str string
	if m.MType == "gauge" {
		str = fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)
	} else {
		str = fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)
	}
	return hash.Sign(key, str)
}

func (m *Metric) SetHash(key string) {
	m.Hash = m.CalculateHash(key)
}

func (m *Metric) AsText() string {
	var str string
	if m.MType == "gauge" {
		str = fmt.Sprintf("%.3f", *m.Value)
	} else {
		str = fmt.Sprintf("%d", *m.Delta)
	}
	if m.Hash != "" {
		str += fmt.Sprintf(";%s", m.Hash)
	}
	return str
}

type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Hash    string `json:"hash,omitempty"`
}

func (s *Response) CalculateHash(key string) string {
	return hash.Sign(key, fmt.Sprintf("msg:%s;err:%s", s.Message, s.Error))
}

func (s *Response) SetHash(key string) {
	s.Hash = s.CalculateHash(key)
}

func (s *Response) AsText() string {
	var msg string
	if s.Message != "" {
		msg = fmt.Sprintf("meassage:%s;", s.Message)
	}
	if s.Error != "" {
		msg += fmt.Sprintf("error:%s;", s.Error)
	}
	if s.Hash != "" {
		msg += fmt.Sprintf("hash:%s;", s.Hash)
	}
	return msg
}

type MemoryStorage struct {
	counters   map[string]int64
	gauges     map[string]float64
	countersMx sync.RWMutex
	gaugesMx   sync.RWMutex
}

func (s *MemoryStorage) GetGauge(metricID string) (float64, error) {
	s.gaugesMx.RLock()
	v, ok := s.gauges[metricID]
	s.gaugesMx.RUnlock()
	if ok {
		return v, nil
	} else {
		e := fmt.Errorf("counter metric %s does not exists", metricID)
		return -1, e
	}
}

func (s *MemoryStorage) SetGauge(metricID string, metricValue float64) error {
	s.gaugesMx.Lock()
	s.gauges[metricID] = metricValue
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

func (s *MemoryStorage) GetCounter(metricID string) (int64, error) {
	s.countersMx.RLock()
	v, ok := s.counters[metricID]
	s.countersMx.RUnlock()
	if ok {
		return v, nil
	} else {
		e := fmt.Errorf("counter metric %s does not exists", metricID)
		return -1, e
	}
}

func (s *MemoryStorage) IncreaseCounter(metricID string, metricValue int64) error {
	s.countersMx.Lock()
	//S1036 - Unnecessary guard around map access
	s.counters[metricID] += metricValue
	s.countersMx.Unlock()
	return nil
}

func (s *MemoryStorage) UpdateMetrics(metrics []Metric) error {
	for _, m := range metrics {
		switch m.MType {
		case "counter":
			s.countersMx.Lock()
			//S1036 - Unnecessary guard around map access
			s.counters[m.ID] += *m.Delta
			s.countersMx.Unlock()
		case "gauge":
			s.gaugesMx.Lock()
			s.gauges[m.ID] = *m.Value
			s.gaugesMx.Unlock()
		default:
			// we should not be here. All metrics were checked by serializer.DecodeBodyBatch
			log.Printf("ERROR counter %s has unknown metric type: %s", m.ID, m.MType)
		}
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

func (s *MemoryStorage) GetMetrics() ([]Metric, error) {
	var metrics = []Metric{}
	s.countersMx.RLock()
	for k, v := range s.counters {
		metrics = append(metrics, Metric{ID: k, MType: "counter", Delta: &v})
	}
	s.countersMx.RUnlock()

	s.gaugesMx.RLock()
	for k, v := range s.gauges {
		metrics = append(metrics, Metric{ID: k, MType: "gauge", Value: &v})
	}
	s.gaugesMx.RUnlock()
	return metrics, nil

}

type Storage interface {
	GetGauge(metricName string) (float64, error)
	SetGauge(metricName string, metricValue float64) error
	GetAllGauges() (map[string]float64, error)
	GetCounter(metricName string) (int64, error)
	GetMetrics() ([]Metric, error)
	SetCounter(metricName string, metricValue int64) error
	IncreaseCounter(metricName string, metricValue int64) error
	GetAllCounters() (map[string]int64, error)
	ResetCounter(metricName string) error
	Avaliable() error
	Close()
	Init() error
	UpdateMetrics(metrics []Metric) error
}

func NewMemoryStorage() Storage {
	return &MemoryStorage{
		counters:   map[string]int64{},
		gauges:     map[string]float64{},
		gaugesMx:   sync.RWMutex{},
		countersMx: sync.RWMutex{},
	}

}

type ServerResponse interface {
	AsText() string
	SetHash(key string)
}
