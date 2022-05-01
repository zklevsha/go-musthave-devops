package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

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

func getMetric(m serializer.Metrics) (serializer.Metrics, int, error) {
	res := serializer.Metrics{ID: m.ID, MType: m.MType}
	switch m.MType {
	case "counter":
		v, err := storage.Server.GetCounter(m.ID)
		if err != nil {
			e := fmt.Errorf("failed to get  %s: %s", m.ID, err.Error())
			return res, 404, e
		}
		res.Delta = &v
	case "gauge":
		v, err := storage.Server.GetGauge(m.ID)
		if err != nil {
			e := fmt.Errorf("failed to get  %s: %s", m.ID, err.Error())
			return res, 404, e
		}
		res.Value = &v

	}
	return res, http.StatusOK, nil
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	m, statusCode, err := serializer.DecodeBody(r.Body)
	if err != nil {
		w.WriteHeader(statusCode)
		w.Write(serializer.EncodeServerResponse("", err.Error()))
		return
	}
	updateMetric(m)
	w.Write(serializer.EncodeServerResponse("metric was saved", ""))
}

func GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	m, statusCode, err := serializer.DecodeURL(r)
	if err != nil {
		w.WriteHeader(statusCode)
		w.Write(serializer.EncodeServerResponse("", err.Error()))
		return
	}
	result, statusCode, err := getMetric(m)
	if err != nil {
		w.WriteHeader(statusCode)
		w.Write(serializer.EncodeServerResponse("", err.Error()))
		return
	}
	if m.MType == "gauge" {
		w.Write([]byte(fmt.Sprintf("%.3f", *result.Value)))
		return
	}
	w.Write([]byte(fmt.Sprintf("%d", *result.Delta)))
}

func GetMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	m, statusCode, err := serializer.DecodeBody(r.Body)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}

	result, statusCode, err := getMetric(m)
	if err != nil {
		w.WriteHeader(statusCode)
		w.Write(serializer.EncodeServerResponse("", err.Error()))
		return
	}
	json.NewEncoder(w).Encode(result)

}
