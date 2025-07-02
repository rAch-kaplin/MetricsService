package server

import (
	"testing"
	"net/http/httptest"
	"net/http"

	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memStorage"
)

func TestMainHandle(t *testing.T) {
	storage := ms.NewMemStorage()
	handler := MainHandle(storage)

	tests := []struct {
		name       string
		url        string
		method     string
		wantStatus int
	}{
		{
			name:       "Gauge update (valid)",
			url:        "/update/gauge/test_metric/97.25",
			method:     http.MethodPost,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Counter update (valid)",
			url:        "/update/counter/test_counter/90",
			method:     http.MethodPost,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid method",
			url:        "/update/gauge/test_metric/97.45",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Invalid metric type",
			url:        "/update/invalid/test_metric/80",
			method:     http.MethodPost,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing value",
			url:        "/update/gauge/test_metric/",
			method:     http.MethodPost,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.url, nil)
			defer req.Body.Close()
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			handler(w, req)

			res := w.Result()
			if res.StatusCode != test.wantStatus {
				t.Errorf("expected status %d, got %d", test.wantStatus, res.StatusCode)
			}
		})
	}

}
