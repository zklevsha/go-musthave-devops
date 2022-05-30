package structs

import (
	"fmt"
	"log"
	"net/http"
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

func (s *MemoryStorage) GetMetric(m Metric) (Metric, int, error) {
	switch m.MType {
	case "counter":
		s.countersMx.RLock()
		v, ok := s.counters[m.ID]
		s.countersMx.RUnlock()
		if ok {
			m.Delta = &v
			return m, http.StatusOK, nil
		} else {
			e := fmt.Errorf("counter metric %s does not exists", m.ID)
			log.Printf("ERROR: %s", e.Error())
			return Metric{}, http.StatusNotFound, e
		}
	case "gauge":
		s.gaugesMx.RLock()
		v, ok := s.gauges[m.ID]
		s.gaugesMx.RUnlock()
		if ok {
			m.Value = &v
			return m, http.StatusOK, nil
		} else {
			e := fmt.Errorf("gauge metric %s does not exists", m.ID)
			log.Printf("ERROR: %s", e.Error())
			return Metric{}, http.StatusNotFound, e
		}
	default:
		e := fmt.Errorf("cant get %s. Metric has unknown type: %s", m.ID, m.MType)
		log.Printf("ERROR: %s", e.Error())
		return Metric{}, http.StatusBadRequest, e
	}
}

func (s *MemoryStorage) GetMetrics() ([]Metric, int, error) {
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
	return metrics, http.StatusOK, nil

}

func (s *MemoryStorage) UpdateMetric(m Metric) (int, error) {
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
		e := fmt.Errorf("cant update %s. Metric has unknown type: %s", m.ID, m.MType)
		log.Printf("ERROR: %s", e.Error())
		return http.StatusBadRequest, e
	}
	return http.StatusOK, nil
}

func (s *MemoryStorage) UpdateMetrics(metrics []Metric) (int, error) {
	for _, m := range metrics {
		statusCode, err := s.UpdateMetric(m)
		if err != nil {
			return statusCode, err
		}
	}
	return http.StatusOK, nil
}

func (s *MemoryStorage) ResetCounter(ID string) error {
	if _, ok := s.counters[ID]; ok {
		s.countersMx.Lock()
		s.counters[ID] = 0
		s.countersMx.Unlock()
		return nil
	} else {
		return fmt.Errorf("counter metric %s does not exists", ID)
	}

}

func (s *MemoryStorage) Avaliable() error {
	return nil

}

func (s *MemoryStorage) Close() {
}

func (s *MemoryStorage) Init() error {
	return nil
}

type Storage interface {
	GetMetric(metric Metric) (Metric, int, error)
	GetMetrics() ([]Metric, int, error)
	UpdateMetric(metric Metric) (int, error)
	UpdateMetrics(metrics []Metric) (int, error)
	ResetCounter(ID string) error
	Avaliable() error
	Close()
	Init() error
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
