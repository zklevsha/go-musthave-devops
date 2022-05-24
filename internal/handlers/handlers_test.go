package handlers_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/handlers"
)

// TestUpdateMeticHandler
func TestUpdateMeticHandler(t *testing.T) {
	h := handlers.Handlers{}
	type want struct {
		code     int
		response string
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
				response: "metric was saved",
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
				response: "metric was saved",
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
			router := mux.NewRouter()
			router.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", h.UpdateMeticHandler)
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
