// Package handlers stores http routes (request handle functions)
// @title Monitoring API
// @description Service for storing and retreiving metrics
package handlers

import (
	"bytes"
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/zklevsha/go-musthave-devops/docs"
	"github.com/zklevsha/go-musthave-devops/internal/archive"
	"github.com/zklevsha/go-musthave-devops/internal/config"
	"github.com/zklevsha/go-musthave-devops/internal/rsaencrypt"
	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
)

func GetErrStatusCode(err error) int {
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
	Storage structs.Storage
	key     string
	privKey *rsa.PrivateKey
}

func (h *Handlers) sendResponse(w http.ResponseWriter, r *http.Request, code int,
	resp structs.ServerResponse) {

	compress :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	asText := !strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")
	asHTML := strings.Contains(strings.Join(r.Header["Accept"], ","), "html")
	responseBody, err := serializer.EncodeServerResponse(resp, compress, asText, h.key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to encode server response: %s", err.Error())))
		return
	}

	if compress {
		w.Header().Set("Content-Encoding", "gzip")
	}

	if asHTML {
		w.Header().Set("Content-Type", "text/html")
	} else if asText {
		w.Header().Set("Content-Type", "text/plain")
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.WriteHeader(code)
	w.Write(responseBody)
}

// UpdateMeticHandler godoc
// @Summary  Set/Update metric
// @Description Set or Update metric value
// @Tags metrics
// @Produce  json
// @Produce text/plain
// @Param  metricType path string true "metric type" enums(counter,gauge)
// @Param  metricID path string true "metric id"
// @Param  metricValue path string true "metric value"
// @Success 200 {object} structs.Response
// @Failure 400 {object} structs.Response
// @Failure 501 {object} structs.Response
// @Router /update/{metricType}/{metricID}/{metricValue} [post]
func (h *Handlers) UpdateMeticHandler(w http.ResponseWriter, r *http.Request) {
	m, err := serializer.DecodeURL(r)
	if err != nil {
		h.sendResponse(w, r, GetErrStatusCode(err), &structs.Response{Error: err.Error()})
		return
	}

	if m.MType == "counter" && m.Delta == nil {
		e := "delta attribute is not set"
		h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
		return
	}
	if m.MType == "gauge" && m.Value == nil {
		e := "gauge attribute is not set"
		h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
		return
	}
	if h.key != "" && m.CalculateHash(h.key) != m.Hash {
		h.sendResponse(w, r, http.StatusBadRequest,
			&structs.Response{Error: "invalid hash value"})
		return
	}

	err = h.Storage.UpdateMetric(m)
	if err != nil {
		e := fmt.Sprintf("failed to update metric %s: %s", m.AsText(), err.Error())
		h.sendResponse(w, r, GetErrStatusCode(err),
			&structs.Response{Error: e})
		return
	}

	h.sendResponse(w, r, http.StatusOK,
		&structs.Response{Message: "metric was saved"})
}

// UpdateMetricJSONHandler godoc
// @Summary  Set/Update metric
// @Description Set or Update metrics value
// @Tags metrics
// @Produce  json
// @Produce text/plain
// @Param metrics body structs.Metric true "Metric to set/update"
// @Success 200 {object} structs.Response
// @Failure 400 {object} structs.Response
// @Failure 501 {object} structs.Response
// @Router /update/ [post]
func (h *Handlers) UpdateMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxBody{} should be set in read body middleware
	b := r.Context().Value(structs.RequestCtxBody{}).([]byte)

	m, err := serializer.DecodeBody(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
		return
	}

	if m.MType == "counter" && m.Delta == nil {
		e := "delta attribute is not set"
		h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
	}
	if m.MType == "gauge" && m.Value == nil {
		e := "gauge attribute is not set"
		h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
	}

	if h.key != "" && m.CalculateHash(h.key) != m.Hash {
		h.sendResponse(w, r, http.StatusBadRequest,
			&structs.Response{Error: "invalid hash value"})
		return
	}

	err = h.Storage.UpdateMetric(m)
	if err != nil {
		e := fmt.Sprintf("failed to update metric %s: %s", m.ID, err.Error())
		h.sendResponse(w, r, GetErrStatusCode(err),
			&structs.Response{Error: e})
		return
	}

	h.sendResponse(w, r, http.StatusOK,
		&structs.Response{Message: "metric was saved"})

}

// UpdateMeticsBatchHandler godoc
// @Summary  Set/Update metrics
// @Description Set or Update multiple metrics at once
// @Tags metrics
// @Produce  json
// @Produce text/plain
// @Param metrics body structs.Metrics true "List of metrics to set/update"
// @Success 200 {object} structs.Response{}
// @Failure 400 {object} structs.Response{}
// @Failure 501 {object} structs.Response{}
// @Router /updates/ [post]
func (h *Handlers) UpdateMeticsBatchHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxBody{} should be set in read body middleware
	b := r.Context().Value(structs.RequestCtxBody{}).([]byte)
	metrics, err := serializer.DecodeBodyBatch(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		log.Printf("ERROR %s", e)
		h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
		return
	}
	log.Println("INFO updating metrics batch")
	err = h.Storage.UpdateMetrics(metrics)
	if err != nil {
		e := fmt.Sprintf("failed to update metric batch: %s", err.Error())
		log.Printf("ERROR %s", e)
		h.sendResponse(w, r, GetErrStatusCode(err), &structs.Response{Error: e})
		return
	}
	m := "metrics batch was updated"
	log.Printf("INFO %s", m)
	h.sendResponse(w, r, http.StatusOK, &structs.Response{Message: m})
}

//GetMetricHandler godoc
// @Summary  Get metric
// @Description Retreiving metric value
// @Tags metrics
// @Produce  json
// @Param  metricType path string true "metric type" enums(counter,gauge)
// @Param  metricID path string true "metric id"
// @Success 200 {object} structs.Metric
// @Failure 404 {object} structs.Response
// @Failure 501 {object} structs.Response
// @Router /value/{metricType}/{metricID} [get]
func (h *Handlers) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	m, err := serializer.DecodeURL(r)
	if err != nil {
		e := fmt.Sprintf("failed to decode url: %s", err.Error())
		h.sendResponse(w, r, GetErrStatusCode(err), &structs.Response{Error: e})
		return
	}
	metric, err := h.Storage.GetMetric(m)
	if err != nil {
		log.Printf(" WARN failed to get metric: %s", err.Error())
		h.sendResponse(w, r, GetErrStatusCode(err), &structs.Response{Error: err.Error()})
		return
	}

	h.sendResponse(w, r, http.StatusOK, &metric)
}

//GetMetricJSONHandler godoc
// @Summary Get  metric
// @Description Retreiving metric value
// @Tags metrics
// @Accept json
// @Produce  json
// @Produce text/plain
// @Param metric body structs.MetricGet true "Get value for metric"
// @Success 200 {object} structs.Metric
// @Failure 404 {object} structs.Response
// @Failure 501 {object} structs.Response
// @Router /value/ [post]
func (h *Handlers) GetMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxBody{} should be set in read body middleware
	b := r.Context().Value(structs.RequestCtxBody{}).([]byte)
	m, err := serializer.DecodeBody(bytes.NewReader(b))
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
		return
	}
	metric, err := h.Storage.GetMetric(m)

	if err != nil {
		e := fmt.Sprintf("failed to get metric: %s", err.Error())
		log.Printf("WARN %s", e)
		h.sendResponse(w, r, GetErrStatusCode(err), &structs.Response{Error: e})
		return
	}

	h.sendResponse(w, r, http.StatusOK, &metric)

}

func (h *Handlers) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	resp := &structs.Response{Message: "<html><body><h1>Server is wokring</h1></body></html>"}
	h.sendResponse(w, r, http.StatusOK, resp)
}

// Ping godoc
// @Summary Ping Database
// @Description Checking if DB is available
// @Tags self-health
// @Produce  json
// @Produce text/plain
// @Success 200 {object} structs.Response
// @Router /ping [get]
func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.Storage.Avaliable()
	if err != nil {
		h.sendResponse(w, r, http.StatusInternalServerError,
			&structs.Response{Error: fmt.Sprintf("DB is down: %s", err.Error())})
		return
	} else {
		h.sendResponse(w, r, http.StatusOK,
			&structs.Response{Message: "DB is working correctly"})
		return
	}
}

func (h *Handlers) ReadBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reading bytes
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			e := fmt.Sprintf("failed to read body: %s", err.Error())
			h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
			return
		}

		// Decrypt
		if h.privKey != nil {
			b, err = rsaencrypt.Decrypt(h.privKey, b, []byte(config.RsaLabel))
			if err != nil {
				e := fmt.Sprintf("failed to decrypt body: %s", err.Error())
				h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
				return
			}
		}

		// Decompress
		reqCompressed :=
			strings.Contains(strings.Join(r.Header["Content-Encoding"], ","), "gzip")
		if reqCompressed {
			b, err = archive.Decompress(b)
			if err != nil {
				e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
				h.sendResponse(w, r, http.StatusBadRequest, &structs.Response{Error: e})
				return
			}
		}

		// Adding body to context
		ctx := context.WithValue(r.Context(), structs.RequestCtxBody{}, b)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetHandler(c config.ServerConfig, store structs.Storage, privKey *rsa.PrivateKey) http.Handler {
	r := mux.NewRouter()
	h := Handlers{key: c.Key, Storage: store, privKey: privKey}

	// root
	r.HandleFunc("/", h.RootHandler)

	// update metric (path parameters)
	r.HandleFunc("/update/{metricType}/{metricID}/{metricValue}",
		h.UpdateMeticHandler).Methods("POST")

	// update metric (body parameter)
	chain := h.ReadBodyMiddleware(http.HandlerFunc(h.UpdateMetricJSONHandler))
	r.Handle("/update/", chain).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// get metric (path parameter)
	r.HandleFunc("/value/{metricType}/{metricID}",
		h.GetMetricHandler).Methods("GET")

	// get metric (body parameter)
	chain = h.ReadBodyMiddleware(http.HandlerFunc(h.GetMetricJSONHandler))
	r.Handle("/value/", chain).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// update multiple metrics
	chain = h.ReadBodyMiddleware(http.HandlerFunc(h.UpdateMeticsBatchHandler))
	r.Handle("/updates/", chain).
		Methods("POST").
		Headers("Content-Type", "application/json")

	r.HandleFunc("/ping", h.Ping)

	// Swagger docs avaialble at /swagger/ endpoint
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}
