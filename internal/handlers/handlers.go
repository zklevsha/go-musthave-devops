package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

func saveMetric(metricType string, metricName string, metricValue string) (int, error) {
	switch metricType {
	case "counter":
		i, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			msg := fmt.Errorf("failed to convert %s to int64: %s", metricName, err.Error())
			return http.StatusBadRequest, msg
		}
		storage.Server.IncreaseCounter(metricName, i)
		return http.StatusOK, nil
	case "gauge":
		f, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			e := fmt.Errorf("failed to convert %s to float64: %s", metricName, err.Error())
			return http.StatusBadRequest, e
		}
		storage.Server.SetGauge(metricName, f)
		return http.StatusOK, nil
	default:
		e := fmt.Errorf("unknown metric type %s", metricType)
		return http.StatusNotImplemented, e
	}
}

func getMetric(metricType string, metricName string) (string, int, error) {
	switch metricType {
	case "counter":
		v, err := storage.Server.GetCounter(metricName)
		if err != nil {
			e := fmt.Errorf("failed to get  %s: %s", metricName, err.Error())
			return "", 404, e
		} else {
			return fmt.Sprintf("%d", v), http.StatusOK, nil
		}
	case "gauge":
		v, err := storage.Server.GetGauge(metricName)
		if err != nil {
			e := fmt.Errorf("failed to get  %s: %s", metricName, err.Error())
			return "", 404, e
		} else {
			return fmt.Sprintf("%.3f", v), http.StatusOK, nil
		}
	default:
		e := fmt.Errorf("unknown metric type %s", metricType)
		return "", http.StatusNotImplemented, e
	}

}

func UpdateMeticHandler(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	statusCode, err := saveMetric(
		v["metricType"], v["metricName"], v["metricValue"])
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	} else {
		w.Write([]byte("metric was saved"))
	}
}

func GetMericHandler(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	value, statusCode, err := getMetric(v["metricType"], v["metricName"])
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	} else {
		w.Write([]byte(value))
	}
}
