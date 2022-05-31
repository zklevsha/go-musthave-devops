package structs

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

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

func NewMemoryStorage() Storage {
	return &MemoryStorage{
		counters:   map[string]int64{},
		gauges:     map[string]float64{},
		gaugesMx:   sync.RWMutex{},
		countersMx: sync.RWMutex{},
	}
}
