package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zklevsha/go-musthave-devops/internal/config"
	"github.com/zklevsha/go-musthave-devops/internal/handlers"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
)

// TestUpdateMeticHandler
func TestUpdateMeticHandler(t *testing.T) {
	type want struct {
		response string
		code     int
	}

	type metric struct {
		metricType  string
		metricName  string
		metricValue string
	}

	tt := []struct {
		name   string
		metric metric
		want   want
	}{
		{
			name: "trying to update counter",
			metric: metric{
				metricType:  "counter",
				metricName:  "testCounter",
				metricValue: "1",
			},
			want: want{
				code:     200,
				response: "meassage:metric was saved;",
			},
		},

		{
			name: "trying to update gauge",
			metric: metric{
				metricType:  "gauge",
				metricName:  "testGauge",
				metricValue: "1.5",
			},
			want: want{
				code:     200,
				response: "meassage:metric was saved;",
			},
		},
	}
	for _, tc := range tt {
		// запускаем каждый тест
		t.Run(tc.name, func(t *testing.T) {
			m := tc.metric
			path := fmt.Sprintf("/update/%s/%s/%s",
				m.metricType, m.metricName, m.metricValue)
			req, err := http.NewRequest("POST", path, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			router := handlers.GetHandler(config.ServerConfig{}, structs.NewMemoryStorage(), nil)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != tc.want.code {
				t.Errorf("Expected status code %d, got %d", tc.want.code, rr.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tc.want.response {
				t.Errorf("Expected body %s, got %s", tc.want.response, rr.Body.String())
			}
		})
	}
}

func TestGetErrStatusCode(t *testing.T) {
	tt := []struct {
		err  error
		name string
		want int
	}{
		{
			name: "metric not found error",
			err:  structs.ErrMetricNotFound,
			want: http.StatusNotFound,
		},
	}

	// запускаем каждый тест
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := handlers.GetErrStatusCode(tc.err)
			if res != tc.want {
				t.Errorf("Expected HTTP status %d, got %d",
					tc.want, res)
			}
		})
	}

}

func TestUpdateMeticJSONHandler(t *testing.T) {
	counter := int64(1)
	gauge := float64(1.5)

	type want struct {
		response string
		code     int
	}

	tt := []struct {
		name   string
		metric structs.Metric
		want   want
	}{
		{
			name: "trying to update counter",

			metric: structs.Metric{
				MType: "counter",
				ID:    "testCounter",
				Delta: &counter,
			},
			want: want{
				code:     200,
				response: "meassage:metric was saved;",
			},
		},

		{
			name: "trying to update gauge",
			metric: structs.Metric{
				MType: "gauge",
				ID:    "testGauge",
				Value: &gauge,
			},
			want: want{
				code:     200,
				response: "meassage:metric was saved;",
			},
		},
	}
	for _, tc := range tt {
		// запускаем каждый тест
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.metric)
			if err != nil {
				t.Fatal(err)
			}
			path := "/update/"
			req, err := http.NewRequest("POST", path, bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router := handlers.GetHandler(config.ServerConfig{}, structs.NewMemoryStorage(), nil)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != tc.want.code {
				t.Errorf("Expected status code %d, got %d", tc.want.code, rr.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tc.want.response {
				t.Errorf("Expected body %s, got %s", tc.want.response, rr.Body.String())
			}
		})
	}
}

func TestUpdateMeticsBatchHandler(t *testing.T) {
	counter := int64(1)
	gauge := float64(1.5)

	type want struct {
		response string
		code     int
	}

	tt := []struct {
		want    want
		name    string
		metrics []structs.Metric
	}{
		{
			name: "trying to update counter and gauge",
			metrics: []structs.Metric{
				{
					MType: "counter",
					ID:    "testCounter",
					Delta: &counter,
				},
				{
					MType: "gauge",
					ID:    "testGauge",
					Value: &gauge,
				},
			},
			want: want{
				code:     200,
				response: "meassage:metrics batch was updated;",
			},
		},
	}
	for _, tc := range tt {
		// запускаем каждый тест
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.metrics)
			if err != nil {
				t.Fatal(err)
			}
			path := "/updates/"
			req, err := http.NewRequest("POST", path, bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router := handlers.GetHandler(config.ServerConfig{}, structs.NewMemoryStorage(), nil)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != tc.want.code {
				t.Errorf("Expected status code %d, got %d", tc.want.code, rr.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tc.want.response {
				t.Errorf("Expected body %s, got %s", tc.want.response, rr.Body.String())
			}
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	storage := structs.NewMemoryStorage()
	counter := int64(1)
	gauge := float64(1.5)

	type want struct {
		response string
		code     int
	}

	tt := []struct {
		name   string
		metric structs.Metric
		want   want
	}{
		{
			name: "trying to update counter",

			metric: structs.Metric{
				MType: "counter",
				ID:    "testCounter",
				Delta: &counter,
			},
			want: want{
				code:     200,
				response: fmt.Sprintf("%d", counter),
			},
		},
		{
			name: "trying to update gauge",
			metric: structs.Metric{
				MType: "gauge",
				ID:    "testGauge",
				Value: &gauge,
			},
			want: want{
				code:     200,
				response: fmt.Sprintf("%.3f", gauge),
			},
		},
	}

	for _, tc := range tt {
		// запускаем каждый тест
		t.Run(tc.name, func(t *testing.T) {
			m := tc.metric

			// saving metric in memory storage
			err := storage.UpdateMetric(m)
			if err != nil {
				t.Fatal(err)
			}

			path := fmt.Sprintf("/value/%s/%s", m.MType, m.ID)
			fmt.Println(path)
			req, err := http.NewRequest("GET", path, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router := handlers.GetHandler(config.ServerConfig{}, storage, nil)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != tc.want.code {
				t.Errorf("Expected status code %d, got %d", tc.want.code, rr.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tc.want.response {
				t.Errorf("Expected body %s, got %s", tc.want.response, rr.Body.String())
			}
		})
	}
}

func TestGetMetricJSONHandler(t *testing.T) {
	storage := structs.NewMemoryStorage()
	counter := int64(1)
	gauge := float64(1.5)

	type want struct {
		response string
		code     int
	}
	tt := []struct {
		name   string
		metric structs.Metric
		want   want
	}{
		{
			name: "trying to get counter",
			metric: structs.Metric{
				MType: "counter",
				ID:    "testCounter",
				Delta: &counter,
			},
			want: want{
				code:     200,
				response: fmt.Sprintf("%d", counter),
			},
		},
		{
			name: "trying to get gauge",
			metric: structs.Metric{
				MType: "gauge",
				ID:    "testGauge",
				Value: &gauge,
			},
			want: want{
				code:     200,
				response: fmt.Sprintf("%.3f", gauge),
			},
		},
	}

	for _, tc := range tt {
		// запускаем каждый тест
		t.Run(tc.name, func(t *testing.T) {
			m := tc.metric

			// saving metric in memory storage
			err := storage.UpdateMetric(m)
			if err != nil {
				t.Fatal(err)
			}

			body, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			path := "/value/"
			req, err := http.NewRequest("POST", path, bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router := handlers.GetHandler(config.ServerConfig{}, storage, nil)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != tc.want.code {
				t.Errorf("Expected status code %d, got %d", tc.want.code, rr.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tc.want.response {
				t.Errorf("Expected body %s, got %s", tc.want.response, rr.Body.String())
			}
		})
	}
}

func TestPing(t *testing.T) {

	type want struct {
		response string
		code     int
	}

	tt := []struct {
		name string
		want want
	}{
		{
			name: "DB is working",
			want: want{code: 200, response: "meassage:DB is working correctly;"},
		},
	}

	for _, tc := range tt {
		// запускаем каждый тест
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/ping", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			router := handlers.GetHandler(config.ServerConfig{}, structs.NewMemoryStorage(), nil)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != tc.want.code {
				t.Errorf("Expected status code %d, got %d", tc.want.code, rr.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tc.want.response {
				t.Errorf("Expected body %s, got %s", tc.want.response, rr.Body.String())
			}
		})
	}

}

func TestRootHandler(t *testing.T) {

	type want struct {
		response string
		code     int
	}

	tt := []struct {
		name string
		want want
	}{
		{
			name: "Test root handler",
			want: want{code: 200, response: "Server is wokring"},
		},
	}

	for _, tc := range tt {
		// запускаем каждый тест
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			router := handlers.GetHandler(config.ServerConfig{}, structs.NewMemoryStorage(), nil)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != tc.want.code {
				t.Errorf("Expected status code %d, got %d", tc.want.code, rr.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(resBody), tc.want.response) {
				t.Errorf("Body '%s' does not contains %s", rr.Body.String(), tc.want.response)
			}
		})
	}

}

func TestGetHandler(t *testing.T) {
	name := "testing GetHandler"
	t.Run(name, func(t *testing.T) {
		handlers.GetHandler(config.ServerConfig{}, structs.NewMemoryStorage(), nil)
	})

}
