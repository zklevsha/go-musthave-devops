package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

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
var urlRegexp = regexp.MustCompile(`^\/update\/(counter|gauge)\/[A-Za-z]+\/(\d+(?:\.\d+)?)$`)

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

func pasrseUrl(url string) (metricType string, metricName string, metricValue string, err error) {
	if !urlRegexp.Match([]byte(url)) {
		return "", "", "", errors.New("failed to parse url. Expected format /update/<metric type>/<metric_name>/<metric_value>")
	}
	spl := strings.Split(url, "/")
	return spl[2], spl[3], spl[4], nil
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
	http.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "for sending metrics use POST", http.StatusBadRequest)
			return
		}
		metricType, metricName, metricValue, err := pasrseUrl(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = saveMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	})

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
