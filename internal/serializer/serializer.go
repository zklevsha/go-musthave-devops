package serializer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func DecodeBody(body io.Reader) (Metrics, int, error) {
	var m Metrics
	err := json.NewDecoder(body).Decode(&m)
	if err != nil {
		return Metrics{}, http.StatusBadRequest, err
	}
	return m, http.StatusOK, nil
}

func DecodeURL(r *http.Request) (Metrics, int, error) {
	v := mux.Vars(r)
	metricName := v["metricName"]
	metricType := v["metricType"]
	metricValue := v["metricValue"]
	switch metricType {
	case "counter":
		i, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			e := fmt.Errorf("failed to convert %s to int64: %s", metricName, err.Error())
			return Metrics{}, http.StatusBadRequest, e
		} else {
			m := Metrics{ID: metricName, MType: metricType, Delta: &i}
			return m, http.StatusOK, nil
		}
	case "gauge":
		f, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			e := fmt.Errorf("failed to convert %s to float64: %s", metricName, err.Error())
			return Metrics{}, http.StatusBadRequest, e
		} else {
			m := Metrics{ID: metricName, MType: metricType, Value: &f}
			return m, http.StatusOK, nil
		}
	default:
		e := fmt.Errorf("unknown metric type %s", metricType)
		return Metrics{}, http.StatusNotImplemented, e
	}
}

func EncodeBodyGauge(id string, value float64) ([]byte, error) {
	m := Metrics{ID: id, MType: "gauge", Value: &value}
	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func EncodeBodyCounter(id string, value int64) ([]byte, error) {
	m := Metrics{ID: id, MType: "counter", Delta: &value}
	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return j, nil
}
