package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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

	if compress {
		w.Header().Set("Content-Encoding", "gzip")
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

	requestCompressed :=
		strings.Contains(strings.Join(r.Header["Content-Encoding"], ","), "gzip")
	compressResponse :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, compressResponse)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, compressResponse)
			return
		}
	}

	m, err := serializer.DecodeBody(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("Failed to decode request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, compressResponse)
		return
	}
	updateMetric(m)
	sendResponse(w, http.StatusOK, serializer.ServerResponse{Result: "metric was saved"}, compressResponse)

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
	log.Printf("GetMetricJSONHandler: request header %+v", r.Header)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	requestCompressed :=
		strings.Contains(strings.Join(r.Header["Content-Encoding"], ","), "gzip")
	compressResponse :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, compressResponse)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, compressResponse)
			return
		}
	}

	m, err := serializer.DecodeBody(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, serializer.ServerResponse{Error: e}, compressResponse)
		return
	}

	metric, statusCode, err := getMetric(m)
	if err != nil {
		e := fmt.Sprintf("failed to get metric: %s", err.Error())
		sendResponse(w, statusCode, serializer.ServerResponse{Error: e}, compressResponse)
		return
	}

	response, err := json.Marshal(metric)
	if err != nil {
		e := fmt.Sprintf("failed to convert metrics to json: %s", err.Error())
		sendResponse(w, http.StatusInternalServerError, serializer.ServerResponse{Error: e}, compressResponse)
		return
	}

	if compressResponse {
		response, err = archive.Compress(response)
		if err != nil {
			e := fmt.Sprintf("failed to compress response: %s", err.Error())
			sendResponse(w, http.StatusInternalServerError, serializer.ServerResponse{Error: e}, compressResponse)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.Write(response)

}

func rootHandrer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	var err error
	compressResponse :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")

	resp := []byte("<html><body><h1>Server is wokring</h1></body></html>")
	if compressResponse {
		resp, err = archive.Compress(resp)
		if err != nil {
			e := fmt.Sprintf("failed to compress response: %s", err.Error())
			sendResponse(w, http.StatusInternalServerError, serializer.ServerResponse{Error: e}, compressResponse)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")

	}
	w.Write(resp)
}

func GetHandler() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", rootHandrer)

	r.HandleFunc("/update/{metricType}/{metricID}/{metricValue}",
		UpdateMeticHandler).Methods("POST")

	r.HandleFunc("/update/", UpdateMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	r.HandleFunc("/value/{metricType}/{metricID}",
		GetMetricHandler).Methods("GET")

	r.HandleFunc("/value/", GetMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return r
}
