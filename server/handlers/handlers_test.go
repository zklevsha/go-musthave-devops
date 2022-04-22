package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zklevsha/go-musthave-devops/server/handlers"
)

// TestUpdateMeticHandler
func TestUpdateMeticHandler(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code     int
		response string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "trying to send counter",
			url:  "/update/counter/testCounter/1",
			want: want{
				code:     200,
				response: "metric was saved",
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handlers.UpdateMeticHandler)
			h.ServeHTTP(w, r)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response, w.Body.String())
			}
		})
	}
}
