package structs

import (
	"fmt"

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
