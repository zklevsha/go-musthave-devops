package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/archive"
	"github.com/zklevsha/go-musthave-devops/internal/config"
	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
)

func getErrStatusCode(err error) int {
	switch {
	case errors.Is(err, structs.ErrMetricNotFound):
		return http.StatusNotFound
	case errors.Is(err, structs.ErrMetricBadType):
		return http.StatusNotImplemented
	case errors.Is(err, structs.ErrMetricNullAttr) ||
		errors.Is(err, structs.ErrMetricBadAttrValue):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}

type Handlers struct {
	key     string
	Storage structs.Storage
}

func (h *Handlers) sendResponse(w http.ResponseWriter, code int,
	resp structs.ServerResponse, compress bool, asText bool) {
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

	m, err := serializer.DecodeURL(r)
	if err != nil {
		h.sendResponse(w, getErrStatusCode(err), &structs.Response{Error: err.Error()}, сompress, asText)
		return
	}

	if m.MType == "counter" && m.Delta == nil {
		e := "delta attribute is not set"
		h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
			сompress, asText)
		return
	}
	if m.MType == "gauge" && m.Value == nil {
		e := "gauge attribute is not set"
		h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
			сompress, asText)
		return
	}
	if h.key != "" && m.CalculateHash(h.key) != m.Hash {
		h.sendResponse(w, http.StatusBadRequest,
			&structs.Response{Error: "invalid hash value"},
			сompress, asText)
		return
	}

	err = h.Storage.UpdateMetric(m)
	if err != nil {
		e := fmt.Sprintf("failed to update metric %s: %s", m.AsText(), err.Error())
		h.sendResponse(w, getErrStatusCode(err),
			&structs.Response{Error: e},
			сompress, asText)
		return
	}

	h.sendResponse(w, http.StatusOK,
		&structs.Response{Message: "metric was saved"}, сompress, asText)
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
			&structs.Response{Error: e}, compressResponse, asText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
				compressResponse, asText)
			return
		}
	}

	m, err := serializer.DecodeBody(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
			compressResponse, asText)
		return
	}

	if m.MType == "counter" && m.Delta == nil {
		e := "delta attribute is not set"
		h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
			compressResponse, asText)
	}
	if m.MType == "gauge" && m.Value == nil {
		e := "gauge attribute is not set"
		h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
			compressResponse, asText)
	}

	if h.key != "" && m.CalculateHash(h.key) != m.Hash {
		h.sendResponse(w, http.StatusBadRequest,
			&structs.Response{Error: "invalid hash value"},
			compressResponse, asText)
		return
	}

	err = h.Storage.UpdateMetric(m)
	if err != nil {
		e := fmt.Sprintf("failed to update metric %s: %s", m.ID, err.Error())
		h.sendResponse(w, getErrStatusCode(err),
			&structs.Response{Error: e},
			compressResponse, asText)
		return
	}

	h.sendResponse(w, http.StatusOK, &structs.Response{Message: "metric was saved"},
		compressResponse, asText)

}

func (h *Handlers) UpdateMeticsBatchHandler(w http.ResponseWriter, r *http.Request) {
	requestCompressed :=
		strings.Contains(strings.Join(r.Header["Content-Encoding"], ","), "gzip")
	compressResponse :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	responseAsText :=
		!strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
			compressResponse, responseAsText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
				compressResponse, responseAsText)
			return
		}
	}

	metrics, err := serializer.DecodeBodyBatch(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		log.Printf("ERROR %s", e)
		h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}
	log.Println("INFO updating metrics batch")
	err = h.Storage.UpdateMetrics(metrics)
	if err != nil {
		e := fmt.Sprintf("failed to update metric batch: %s", err.Error())
		log.Printf("ERROR %s", e)
		h.sendResponse(w, getErrStatusCode(err), &structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}
	m := "metrics batch was updated"
	log.Printf("INFO %s", m)
	h.sendResponse(w, http.StatusOK, &structs.Response{Message: m},
		compressResponse, responseAsText)
}

func (h *Handlers) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	сompress :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	asText := !strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	m, err := serializer.DecodeURL(r)
	if err != nil {
		e := fmt.Sprintf("failed to decode url: %s", err.Error())
		h.sendResponse(w, getErrStatusCode(err), &structs.Response{Error: e}, сompress, asText)
		return
	}
	metric, err := h.Storage.GetMetric(m)
	if err != nil {
		log.Printf(" WARN failed to get metric: %s", err.Error())
		h.sendResponse(w, getErrStatusCode(err), &structs.Response{Error: err.Error()}, сompress, asText)
		return
	}

	h.sendResponse(w, http.StatusOK, &metric, сompress, asText)
}

func (h *Handlers) GetMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
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
		h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
			compressResponse, responseAsText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
				compressResponse, responseAsText)
			return
		}
	}

	m, err := serializer.DecodeBody(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		h.sendResponse(w, http.StatusBadRequest, &structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}
	metric, err := h.Storage.GetMetric(m)

	if err != nil {
		e := fmt.Sprintf("failed to get metric: %s", err.Error())
		log.Printf("WARN %s", e)
		h.sendResponse(w, getErrStatusCode(err), &structs.Response{Error: e},
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

	resp := &structs.Response{Message: "<html><body><h1>Server is wokring</h1></body></html>"}
	h.sendResponse(w, http.StatusOK, resp, compress, asText)
}

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	compress :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	asText := !strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	err := h.Storage.Avaliable()
	if err != nil {
		h.sendResponse(w, http.StatusInternalServerError,
			&structs.Response{Error: fmt.Sprintf("DB is down: %s", err.Error())},
			compress, asText)
		return
	} else {
		h.sendResponse(w, http.StatusOK,
			&structs.Response{Message: "DB is working correctly"},
			compress, asText)
		return
	}
}

func GetHandler(c config.ServerConfig, ctx context.Context, store structs.Storage) http.Handler {
	r := mux.NewRouter()
	h := Handlers{key: c.Key, Storage: store}
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

	r.HandleFunc("/updates/", h.UpdateMeticsBatchHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	r.HandleFunc("/ping", h.Ping)

	return r
}
