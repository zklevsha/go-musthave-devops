package structs

import (
	"log"
	"sync"
)

type MemoryStorage struct {
	counters   map[string]int64
	gauges     map[string]float64
	countersMx sync.RWMutex
	gaugesMx   sync.RWMutex
}

func (s *MemoryStorage) GetMetric(m Metric) (Metric, error) {
	switch m.MType {
	case "counter":
		s.countersMx.RLock()
		v, ok := s.counters[m.ID]
		s.countersMx.RUnlock()
		if ok {
			m.Delta = &v
			return m, nil
		} else {
			log.Printf("WARN: counter metric %s was not found", m.ID)
			return Metric{}, ErrMetricNotFound
		}
	case "gauge":
		s.gaugesMx.RLock()
		v, ok := s.gauges[m.ID]
		s.gaugesMx.RUnlock()
		if ok {
			m.Value = &v
			return m, nil
		} else {
			log.Printf("WARN: gauge metric %s does not exists", m.ID)
			return Metric{}, ErrMetricNotFound
		}
	default:
		log.Printf("WARN:cant get %s. Metric has unknown type: %s", m.ID, m.MType)
		return Metric{}, ErrMetricBadType
	}
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

func (s *MemoryStorage) UpdateMetric(m Metric) error {
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
		log.Printf("ERROR: cant update %s. Metric has unknown type: %s", m.ID, m.MType)
		return ErrMetricBadType
	}
	return nil
}

func (s *MemoryStorage) UpdateMetrics(metrics []Metric) error {
	for _, m := range metrics {
		err := s.UpdateMetric(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryStorage) ResetCounter(ID string) error {
	if _, ok := s.counters[ID]; ok {
		s.countersMx.Lock()
		s.counters[ID] = 0
		s.countersMx.Unlock()
		return nil
	} else {
		return ErrMetricNotFound
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
