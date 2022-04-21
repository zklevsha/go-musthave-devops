package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

const serverSocket = "127.0.0.1:8080"

var counters = map[string]int64{
	"PollCount": 0,
}
var gauges = map[string]float64{
	"Alloc":         0,
	"BuckHashSys":   0,
	"Frees":         0,
	"GCCPUFraction": 0,
	"GCSys":         0,
	"HeapAlloc":     0,
	"HeapIdle":      0,
	"HeapInuse":     0,
	"HeapObjects":   0,
	"HeapReleased":  0,
	"HeapSys":       0,
	"LastGC":        0,
	"Lookups":       0,
	"MCacheInuse":   0,
	"MCacheSys":     0,
	"MSpanInuse":    0,
	"MSpanSys":      0,
	"Mallocs":       0,
	"NextGC":        0,
	"NumForcedGC":   0,
	"NumGC":         0,
	"OtherSys":      0,
	"PauseTotalNs":  0,
	"StackInuse":    0,
	"StackSys":      0,
	"Sys":           0,
	"TotalAlloc":    0,
	"RandomValue":   0,
}

var counterMx sync.Mutex
var gaugeMx sync.Mutex

func saveCounter(metricName string, metricValue int64) error {

	if _, ok := counters[metricName]; ok {
		counterMx.Lock()
		counters[metricName] += metricValue
		counterMx.Unlock()
		return nil
	}
	return fmt.Errorf("counter metric %s does not exists", metricName)
}

func saveGauge(metricName string, metricValue float64) error {
	if _, ok := gauges[metricName]; ok {
		gaugeMx.Lock()
		counters[metricName] = int64(metricValue)
		gaugeMx.Unlock()
		return nil
	}
	return fmt.Errorf("counter metric %s does not exists", metricName)
}

func saveMetric(metricType string, metricName string, metricValue string) error {
	switch metricType {
	case "counter":
		i, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to int64: %s", metricName, err.Error())
		}
		err = saveCounter(metricName, i)
		if err != nil {
			return fmt.Errorf("failed to save %s:  %s", metricName, err.Error())
		}
		return nil
	case "gauge":
		f, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to float64: %s", metricName, err.Error())
		}
		err = saveGauge(metricName, f)
		if err != nil {
			return fmt.Errorf("failed to save %s:  %s", metricName, err.Error())
		}
		return nil
	default:
		return fmt.Errorf("unknown metric type %s", metricType)
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
		err := saveMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			w.Write([]byte("metric was saved"))
		}
	})

	fmt.Printf("Starting server at %s\n", serverSocket)
	if err := http.ListenAndServe(serverSocket, r); err != nil {
		fmt.Println(err.Error())
	}
}
