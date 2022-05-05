package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/archive"
	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

func updateMetric(m serializer.Metric) {
	switch m.MType {
	case "counter":
		log.Printf("INFO updating metric: id:%s, type:counter, delta:%d \n",
			m.ID, *m.Delta)
		storage.Server.IncreaseCounter(m.ID, *m.Delta)
	case "gauge":
		log.Printf("INFO updating metric: id:%s, type:gauge, value:%f \n",
			m.ID, *m.Value)
		storage.Server.SetGauge(m.ID, *m.Value)
	}
}

func sendResponse(w http.ResponseWriter, code int, resp serializer.ServerResponse, compress bool) {
	responseBody, err := serializer.EncodeServerResponse(resp, compress)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to encode server response: %s", err.Error())))
		return
	}
	w.WriteHeader(code)
	w.Write(responseBody)
}

func getMetric(m serializer.Metric) (serializer.Metric, int, error) {
	res := serializer.Metric{ID: m.ID, MType: m.MType}
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
	m, err := serializer.DecodeBody(r.Body)
	if err != nil {
		e := fmt.Sprintf("Failed to decode request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, false)
		return
	}
	updateMetric(m)
	sendResponse(w, http.StatusOK, serializer.ServerResponse{Result: "metric was saved"}, false)
}

func UpdateMeticCompressedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip")
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, true)
	}
	data, err := archive.Decompress(compressed)
	if err != nil {
		e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, true)
		return
	}
	m, err := serializer.DecodeBody(bytes.NewReader(data))
	if err != nil {
		e := fmt.Sprintf("Failed to decode request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, true)
		return
	}
	updateMetric(m)
	sendResponse(w, http.StatusOK, serializer.ServerResponse{Result: "metric was saved"}, true)

}

func GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	m, statusCode, err := serializer.DecodeURL(r)
	if err != nil {
		e := fmt.Sprintf("failed to decode url: %s", err.Error())
		sendResponse(w, statusCode, serializer.ServerResponse{Error: e}, false)
		return
	}
	result, statusCode, err := getMetric(m)
	if err != nil {
		sendResponse(w, statusCode, serializer.ServerResponse{Error: err.Error()}, false)
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
	m, err := serializer.DecodeBody(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, false)
		return
	}
	log.Printf("GetMetricJSONHandler metric: %+v\n", m)

	result, statusCode, err := getMetric(m)
	if err != nil {
		e := fmt.Sprintf("failed to get metric: %s", err.Error())
		sendResponse(w, statusCode, serializer.ServerResponse{Error: e}, false)
		return
	}
	json.NewEncoder(w).Encode(result)

}

func GetMeticCompressedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip")

	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, true)
	}

	data, err := archive.Decompress(compressed)
	if err != nil {
		e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, true)
		return
	}

	m, err := serializer.DecodeBody(bytes.NewReader(data))
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, true)
		return
	}
	log.Printf("GetMetricJSONHandler metric: %+v\n", m)

	metric, statusCode, err := getMetric(m)
	if err != nil {
		e := fmt.Sprintf("failed to get metric: %s", err.Error())
		sendResponse(w, statusCode, serializer.ServerResponse{Error: e}, true)
		return
	}

	j, err := json.Marshal(metric)
	if err != nil {
		e := fmt.Sprintf("failed to convert metrics to json: %s", err.Error())
		sendResponse(w, http.StatusInternalServerError, serializer.ServerResponse{Error: e}, false)
		return
	}

	compressed, err = archive.Compress(j)
	if err != nil {
		e := fmt.Sprintf("failed to compress response: %s", err.Error())
		sendResponse(w, http.StatusInternalServerError, serializer.ServerResponse{Error: e}, false)
		return
	}
	w.Write(compressed)

}

func GetHandler() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/update/{metricType}/{metricID}/{metricValue}",
		UpdateMeticHandler).Methods("POST")

	r.HandleFunc("/update/", UpdateMeticCompressedHandler).
		Methods("POST").
		Headers("Content-Type", "application/json",
			"Content-Encoding", "gzip")

	r.HandleFunc("/update/", UpdateMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	r.HandleFunc("/value/{metricType}/{metricID}",
		GetMetricHandler).Methods("GET")

	r.HandleFunc("/value/", GetMeticCompressedHandler).
		Methods("POST").
		Headers("Content-Type", "application/json",
			"Accept-Encoding", "gzip")

	r.HandleFunc("/value/", GetMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return r
}
