package serializer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/archive"
	"github.com/zklevsha/go-musthave-devops/internal/hash"
	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // hmac метрики
}

func (m *Metric) CalculateHash(key string) string {
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

type Metrics []Metric

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

func (s Response) AsText() string {
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

type ServerResponse interface {
	AsText() string
	SetHash(key string)
}

func DecodeBody(body io.Reader) (Metric, error) {
	var m Metric
	err := json.NewDecoder(body).Decode(&m)
	if err != nil {
		return Metric{}, err
	}
	return m, err
}

func DecodeURL(r *http.Request) (Metric, int, error) {
	v := mux.Vars(r)
	metricID := v["metricID"]
	metricType := v["metricType"]
	metricValue := v["metricValue"]

	if len(metricValue) == 0 {
		return Metric{ID: metricID, MType: metricType}, 200, nil
	}
	switch metricType {
	case "counter":
		i, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			e := fmt.Errorf("failed to convert %s (%s) to int64: %s", metricID, metricValue, err.Error())
			return Metric{}, http.StatusBadRequest, e
		} else {
			m := Metric{ID: metricID, MType: metricType, Delta: &i}
			return m, http.StatusOK, nil
		}
	case "gauge":
		f, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			e := fmt.Errorf("failed to convert %s (%s) to float64: %s", metricID, metricValue, err.Error())
			return Metric{}, http.StatusBadRequest, e
		} else {
			m := Metric{ID: metricID, MType: metricType, Value: &f}
			return m, http.StatusOK, nil
		}
	default:
		e := fmt.Errorf("unknown metric type %s", metricType)
		return Metric{}, http.StatusNotImplemented, e
	}
}

func EncodeBodyGauge(id string, value float64, key string) ([]byte, error) {
	m := Metric{ID: id, MType: "gauge", Value: &value}
	if key != "" {
		m.SetHash(key)
	}
	return json.Marshal(m)
}

func EncodeBodyCounter(id string, value int64, key string) ([]byte, error) {
	m := Metric{ID: id, MType: "counter", Delta: &value}
	if key != "" {
		m.SetHash(key)

	}
	return json.Marshal(m)
}

func EncodeServerResponse(resp ServerResponse, compress bool, asText bool, key string) ([]byte, error) {
	if key != "" {
		resp.SetHash(key)
	}

	var msg []byte
	var err error

	if asText {
		msg = []byte(resp.AsText())
	} else {
		msg, err = json.Marshal(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to encode server response to json %s", err.Error())
		}
	}

	if !compress {
		return msg, nil
	}

	compressed, err := archive.Compress(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to compress server response %s", err.Error())
	}
	return compressed, nil
}

func EncodeMetrics() ([]byte, error) {
	metrics := Metrics{}
	counters := storage.Server.GetAllCounters()
	gauges := storage.Server.GetAllGauges()
	for k := range counters {
		d := counters[k]
		metrics = append(metrics, Metric{ID: k, Delta: &d, MType: "counter"})
	}
	for k := range gauges {
		v := gauges[k]
		metrics = append(metrics, Metric{ID: k, MType: "gauge", Value: &v})
	}

	json, err := json.Marshal(metrics)
	if err != nil {
		return []byte{}, err
	}
	return json, nil
}
