package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

func updateMetric(m serializer.Metrics) {
	switch m.MType {
	case "counter":
		storage.Server.IncreaseCounter(m.ID, *m.Delta)
	case "gauge":
		storage.Server.SetGauge(m.ID, *m.Value)
	}
}

func getMetric(metricType string, metricName string) (string, int, error) {
	switch metricType {
	case "counter":
		v, err := storage.Server.GetCounter(metricName)
		if err != nil {
			e := fmt.Errorf("failed to get  %s: %s", metricName, err.Error())
			return "", 404, e
		} else {
			return fmt.Sprintf("%d", v), http.StatusOK, nil
		}
	case "gauge":
		v, err := storage.Server.GetGauge(metricName)
		if err != nil {
			e := fmt.Errorf("failed to get  %s: %s", metricName, err.Error())
			return "", 404, e
		} else {
			return fmt.Sprintf("%.3f", v), http.StatusOK, nil
		}
	default:
		e := fmt.Errorf("unknown metric type %s", metricType)
		return "", http.StatusNotImplemented, e
	}

}

func UpdateMeticHandler(w http.ResponseWriter, r *http.Request) {
	m, statusCode, err := serializer.DecodeURL(r)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}
	updateMetric(m)
	w.Write([]byte("metric was saved"))
}

func UpdateMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	m, statusCode, err := serializer.DecodeBody(r.Body)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		msg := string(serializer.EncodeServerResponse("", err.Error()))
		http.Error(w, msg, statusCode)

	}
	updateMetric(m)
	w.Write(serializer.EncodeServerResponse("metric was saved", ""))
}

func GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	value, statusCode, err := getMetric(v["metricType"], v["metricName"])
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	} else {
		w.Write([]byte(value))
	}
}

func GetMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	m, statusCode, err := serializer.DecodeBody(r.Body)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}
	value, statusCode, err := getMetric(m.MType, m.ID)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	} else {
		w.Write([]byte(value))
	}
}
