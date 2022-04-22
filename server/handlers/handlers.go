package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/zklevsha/go-musthave-devops/server/storage"
)

func saveCounter(metricName string, metricValue int64) error {

	if _, ok := storage.Counters[metricName]; ok {
		storage.CounterMx.Lock()
		storage.Counters[metricName] += metricValue
		storage.CounterMx.Unlock()
	} else {
		storage.CounterMx.Lock()
		storage.Counters[metricName] = metricValue
		storage.CounterMx.Unlock()
	}
	return nil
}

func saveGauge(metricName string, metricValue float64) error {
	storage.GaugeMx.Lock()
	storage.Gauges[metricName] = metricValue
	storage.GaugeMx.Unlock()
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
		storage.CounterMx.Lock()
		v, ok := storage.Counters[metricName]
		storage.CounterMx.Unlock()
		if ok {
			return fmt.Sprintf("%d", v), http.StatusOK, nil
		} else {
			e := fmt.Errorf("counter metric %s does not exists", metricName)
			return "", 404, e
		}
	case "gauge":
		storage.GaugeMx.Lock()
		v, ok := storage.Gauges[metricName]
		storage.GaugeMx.Lock()
		if ok {
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

func UpdateMeticHandler(w http.ResponseWriter, r *http.Request) {
	split := strings.Split(r.URL.Path, "/")[2:]
	metricType := split[0]
	metricName := split[1]
	metricValue := split[2]
	statusCode, err := saveMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	} else {
		w.Write([]byte("metric was saved"))
	}
}

func GetMericHandler(w http.ResponseWriter, r *http.Request) {
	split := strings.Split(r.URL.Path, "/")[2:]
	metricType := split[0]
	metricName := split[1]
	value, statusCode, err := getMetric(metricType, metricName)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	} else {
		w.Write([]byte(value))
	}
}
