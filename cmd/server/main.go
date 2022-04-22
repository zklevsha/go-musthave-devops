package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

const serverSocket = ":8080"

var counters = make(map[string]int64)

var gauges = make(map[string]float64)

var counterMx sync.Mutex
var gaugeMx sync.Mutex

func saveCounter(metricName string, metricValue int64) error {

	if _, ok := counters[metricName]; ok {
		counterMx.Lock()
		counters[metricName] += metricValue
		counterMx.Unlock()
	} else {
		counterMx.Lock()
		counters[metricName] = metricValue
		counterMx.Unlock()
	}
	return nil
}

func saveGauge(metricName string, metricValue float64) error {
	gaugeMx.Lock()
	gauges[metricName] = metricValue
	gaugeMx.Unlock()
	return nil
}

func saveMetric(metricType string, metricName string, metricValue string) (int, error) {
	switch metricType {
	case "counter":
		i, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			msg := fmt.Errorf("failed to convert %s to int64: %s", metricName, err.Error())
			return http.StatusBadRequest, msg
		}
		err = saveCounter(metricName, i)
		if err != nil {
			e := fmt.Errorf("failed to save %s:  %s", metricName, err.Error())
			return http.StatusBadRequest, e
		}
		return http.StatusOK, nil
	case "gauge":
		f, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			e := fmt.Errorf("failed to convert %s to float64: %s", metricName, err.Error())
			return http.StatusBadRequest, e
		}
		err = saveGauge(metricName, f)
		if err != nil {
			e := fmt.Errorf("failed to save %s:  %s", metricName, err.Error())
			return http.StatusBadRequest, e
		}
		return http.StatusOK, nil
	default:
		e := fmt.Errorf("unknown metric type %s", metricType)
		return http.StatusNotImplemented, e
	}
}

func getMetric(metricType string, metricName string) (string, int, error) {
	switch metricType {
	case "counter":
		if _, ok := counters[metricName]; ok {
			counterMx.Lock()
			v := counters[metricName]
			counterMx.Unlock()
			return fmt.Sprintf("%d", v), http.StatusOK, nil
		} else {
			e := fmt.Errorf("counter metric %s does not exists", metricName)
			return "", 404, e
		}
	case "gauge":
		if _, ok := gauges[metricName]; ok {
			gaugeMx.Lock()
			v := gauges[metricName]
			gaugeMx.Unlock()
			return fmt.Sprintf("%f", v), http.StatusOK, nil
		} else {
			e := fmt.Errorf("gauge metric %s does not exists", metricName)
			return "", 404, e
		}
	default:
		e := fmt.Errorf("unknown metric type %s", metricType)
		return "", http.StatusNotImplemented, e
	}

}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		statusCode, err := saveMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		} else {
			w.Write([]byte("metric was saved"))
		}
	})

	r.Get("/value/{metricType}/{metricName}", func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		value, statusCode, err := getMetric(metricType, metricName)
		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		} else {
			w.Write([]byte(value))
		}
	})

	fmt.Printf("Starting server at %s\n", serverSocket)
	if err := http.ListenAndServe(serverSocket, r); err != nil {
		fmt.Println(err.Error())
	}
}
