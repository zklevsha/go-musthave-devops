package handlers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/archive"
	"github.com/zklevsha/go-musthave-devops/internal/config"
	"github.com/zklevsha/go-musthave-devops/internal/db"
	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

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

type Handlers struct {
	key   string
	useDB bool
	db    db.DbConnector
}

func (h *Handlers) sendResponse(w http.ResponseWriter, code int,
	resp serializer.ServerResponse, compress bool, asText bool) {
	responseBody, err := serializer.EncodeServerResponse(resp, compress, asText, h.key)
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

func (h *Handlers) UpdateMeticHandler(w http.ResponseWriter, r *http.Request) {
	сompress :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	asText := !strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	m, statusCode, err := serializer.DecodeURL(r)
	if err != nil {
		h.sendResponse(w, statusCode, &serializer.Response{Error: err.Error()}, сompress, asText)
		return
	}

	if h.key != "" && m.CalculateHash(h.key) != m.Hash {
		h.sendResponse(w, http.StatusBadRequest,
			&serializer.Response{Error: "invalid hash value"},
			сompress, asText)
		return
	}

	updateMetric(m)

	h.sendResponse(w, http.StatusOK,
		&serializer.Response{Message: "metric was saved"}, сompress, asText)
}

func (h *Handlers) UpdateMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	requestCompressed :=
		strings.Contains(strings.Join(r.Header["Content-Encoding"], ","), "gzip")
	compressResponse :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	asText :=
		!strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		h.sendResponse(w, http.StatusBadRequest,
			&serializer.Response{Error: e}, compressResponse, asText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			h.sendResponse(w, http.StatusBadRequest, &serializer.Response{Error: e},
				compressResponse, asText)
			return
		}
	}

	m, err := serializer.DecodeBody(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("Failed to decode request body: %s", err.Error())
		h.sendResponse(w, http.StatusBadRequest, &serializer.Response{Error: e},
			compressResponse, asText)
		return
	}

	if h.key != "" && m.CalculateHash(h.key) != m.Hash {
		h.sendResponse(w, http.StatusBadRequest,
			&serializer.Response{Error: "invalid hash value"},
			compressResponse, asText)
		return
	}

	updateMetric(m)
	h.sendResponse(w, http.StatusOK, &serializer.Response{Message: "metric was saved"},
		compressResponse, asText)

}

func (h *Handlers) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	сompress :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	asText := !strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	m, statusCode, err := serializer.DecodeURL(r)
	if err != nil {
		e := fmt.Sprintf("failed to decode url: %s", err.Error())
		h.sendResponse(w, statusCode, &serializer.Response{Error: e}, сompress, asText)
		return
	}
	metric, statusCode, err := getMetric(m)
	if err != nil {
		h.sendResponse(w, statusCode, &serializer.Response{Error: err.Error()}, сompress, asText)
		return
	}

	h.sendResponse(w, http.StatusOK, &metric, сompress, asText)
}

func (h *Handlers) GetMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetMetricJSONHandler: request header %+v", r.Header)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	requestCompressed :=
		strings.Contains(strings.Join(r.Header["Content-Encoding"], ","), "gzip")
	compressResponse :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	responseAsText :=
		!strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		h.sendResponse(w, http.StatusBadRequest, &serializer.Response{Error: e},
			compressResponse, responseAsText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			h.sendResponse(w, http.StatusBadRequest, &serializer.Response{Error: e},
				compressResponse, responseAsText)
			return
		}
	}

	m, err := serializer.DecodeBody(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		h.sendResponse(w, http.StatusBadRequest, &serializer.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	metric, statusCode, err := getMetric(m)
	if err != nil {
		e := fmt.Sprintf("failed to get metric: %s", err.Error())
		h.sendResponse(w, statusCode, &serializer.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	h.sendResponse(w, http.StatusOK, &metric, compressResponse, responseAsText)

}

func (h *Handlers) rootHandrer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	compress :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	asText := !strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	resp := &serializer.Response{Message: "<html><body><h1>Server is wokring</h1></body></html>"}
	h.sendResponse(w, http.StatusOK, resp, compress, asText)
}

func (h *Handlers) PingDB(w http.ResponseWriter, r *http.Request) {
	compress :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	asText := !strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	if !h.useDB {
		h.sendResponse(w, http.StatusInternalServerError,
			&serializer.Response{Error: "DB backend is not enabled"},
			compress, asText)
		return
	}

	err := h.db.Avaliable()
	if err != nil {
		h.sendResponse(w, http.StatusInternalServerError,
			&serializer.Response{Error: fmt.Sprintf("DB is down: %s", err.Error())},
			compress, asText)
		return
	} else {
		h.sendResponse(w, http.StatusOK,
			&serializer.Response{Message: "DB is working correctly"},
			compress, asText)
		return
	}
}

func GetHandler(c config.ServerConfig) http.Handler {
	r := mux.NewRouter()
	h := Handlers{key: c.Key, useDB: c.UseDB, db: db.DbConnector{DSN: c.DSN}}
	r.HandleFunc("/", h.rootHandrer)

	r.HandleFunc("/update/{metricType}/{metricID}/{metricValue}",
		h.UpdateMeticHandler).Methods("POST")

	r.HandleFunc("/update/", h.UpdateMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	r.HandleFunc("/value/{metricType}/{metricID}",
		h.GetMetricHandler).Methods("GET")

	r.HandleFunc("/value/", h.GetMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	r.HandleFunc("/ping", h.PingDB)
	return r
}
