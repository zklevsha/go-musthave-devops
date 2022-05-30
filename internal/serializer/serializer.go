package serializer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/archive"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
)

func DecodeBody(body io.Reader) (structs.Metric, error) {
	var m structs.Metric
	err := json.NewDecoder(body).Decode(&m)
	if err != nil {
		return structs.Metric{}, err
	}
	if m.MType != "counter" && m.MType != "gauge" {
		err = fmt.Errorf("uknown metric type: %s", m.MType)
		return m, err
	}

	return m, err
}

func DecodeBodyBatch(body io.Reader) ([]structs.Metric, error) {
	var metrics = []structs.Metric{}
	err := json.NewDecoder(body).Decode(&metrics)
	if err != nil {
		return []structs.Metric{}, fmt.Errorf("cant unmarshal body to []Metrics: %s", err.Error())
	}
	// Data checking
	for _, m := range metrics {
		if m.MType != "counter" && m.MType != "gauge" {
			return []structs.Metric{}, fmt.Errorf("metric %s has unknown type: %s", m.ID, m.MType)
		}
		if m.MType == "counter" && m.Delta == nil {
			return []structs.Metric{}, fmt.Errorf("delta attirbute is not set for counter %s", m.ID)
		}
		if m.MType == "gauge" && m.Value == nil {
			return []structs.Metric{}, fmt.Errorf("value attribute is not set for gauge %s", m.ID)
		}
	}
	return metrics, nil
}

func DecodeURL(r *http.Request) (structs.Metric, int, error) {
	v := mux.Vars(r)
	metricID := v["metricID"]
	metricType := v["metricType"]
	metricValue := v["metricValue"]

	if len(metricValue) == 0 {
		return structs.Metric{ID: metricID, MType: metricType}, 200, nil
	}
	switch metricType {
	case "counter":
		i, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			e := fmt.Errorf("failed to convert %s (%s) to int64: %s", metricID, metricValue, err.Error())
			return structs.Metric{}, http.StatusBadRequest, e
		} else {
			m := structs.Metric{ID: metricID, MType: metricType, Delta: &i}
			return m, http.StatusOK, nil
		}
	case "gauge":
		f, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			e := fmt.Errorf("failed to convert %s (%s) to float64: %s", metricID, metricValue, err.Error())
			return structs.Metric{}, http.StatusBadRequest, e
		} else {
			m := structs.Metric{ID: metricID, MType: metricType, Value: &f}
			return m, http.StatusOK, nil
		}
	default:
		e := fmt.Errorf("unknown metric type %s", metricType)
		return structs.Metric{}, http.StatusNotImplemented, e
	}
}

func EncodeBodyMetrics(metrics []structs.Metric, key string) ([]byte, error) {
	if key != "" {
		for _, m := range metrics {
			m.SetHash(key)
		}
	}
	return json.Marshal(metrics)
}

func EncodeBodyGauge(id string, value float64, key string) ([]byte, error) {
	m := structs.Metric{ID: id, MType: "gauge", Value: &value}
	if key != "" {
		m.SetHash(key)
	}
	return json.Marshal(m)
}

func EncodeBodyCounter(id string, value int64, key string) ([]byte, error) {
	m := structs.Metric{ID: id, MType: "counter", Delta: &value}
	if key != "" {
		m.SetHash(key)

	}
	return json.Marshal(m)
}

func EncodeServerResponse(resp structs.ServerResponse, compress bool, asText bool, key string) ([]byte, error) {
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

func EncodeMetrics(store structs.Storage) ([]byte, error) {
	metrics, _, err := store.GetMetrics()
	if err != nil {
		e := fmt.Errorf("failed to get metrics: %s", err.Error())
		return []byte{}, e
	}
	json, err := json.Marshal(metrics)
	if err != nil {
		return []byte{}, err
	}
	return json, nil
}
