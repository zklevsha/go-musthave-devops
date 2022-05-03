package serializer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type Metrics []Metric

type ServerResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

func DecodeBody(body io.Reader) (Metric, int, error) {
	var m Metric
	err := json.NewDecoder(body).Decode(&m)
	if err != nil {
		return Metric{}, http.StatusBadRequest, err
	}
	return m, http.StatusOK, nil
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

func EncodeBodyGauge(id string, value float64) ([]byte, error) {
	return json.Marshal(Metric{ID: id, MType: "gauge", Value: &value})
}

func EncodeBodyCounter(id string, value int64) ([]byte, error) {
	return json.Marshal(Metric{ID: id, MType: "counter", Delta: &value})
}

func EncodeServerResponse(result string, errorMessage string) []byte {
	j, err := json.Marshal(ServerResponse{result, errorMessage})
	if err != nil {
		return []byte(fmt.Sprintf(`{"result": %s, "error": %s}`, result, errorMessage))
	}
	return j
}

func EncodeMetrics(m Metrics) ([]byte, error) {
	return json.MarshalIndent(m, "", " ")
}
