package server_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	repo "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/repository"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/router"
	srvUsecase "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/server"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetric(t *testing.T) {
	opts := &config.Options{}
	for _, opt := range []func(*config.Options){
		config.WithAddress("localhost:8080"),
		config.WithStoreInterval(300),
		config.WithFileStoragePath("/tmp/metrics-db.json"),
		config.WithRestoreOnStart(true),
	} {
		opt(opts)
	}

	storage := repo.NewMemStorage()
	metricUsecase := srvUsecase.NewMetricUsecase(storage, storage, storage)
	router := router.NewRouter(server.NewServer(metricUsecase, nil))

	tests := []struct {
		name       string
		method     string
		url        string
		wantStatus int
	}{
		{
			name:       "Valid Gauge Update",
			method:     http.MethodPost,
			url:        "/update/gauge/testGauge/123.45",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Valid Counter Update",
			method:     http.MethodPost,
			url:        "/update/counter/testCounter/100",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Update Counter with another value",
			method:     http.MethodPost,
			url:        "/update/counter/testCounter/50",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid Gauge Value",
			method:     http.MethodPost,
			url:        "/update/gauge/invalidGauge/abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid Counter Value",
			method:     http.MethodPost,
			url:        "/update/counter/invalidCounter/xyz",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Unknown Metric Type",
			method:     http.MethodPost,
			url:        "/update/unknown/testMetric/123",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing Metric Name",
			method:     http.MethodPost,
			url:        "/update/gauge//123.45",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestGetMetric(t *testing.T) {
	opts := &config.Options{}
	for _, opt := range []func(*config.Options){
		config.WithAddress("localhost:8080"),
		config.WithStoreInterval(300),
		config.WithFileStoragePath("/tmp/metrics-db.json"),
		config.WithRestoreOnStart(true),
	} {
		opt(opts)
	}

	//FIXME - maybe need mocks
	ctx := context.Background()
	storage := repo.NewMemStorage()

	if err := storage.UpdateMetric(ctx, models.GaugeType, "cpu_usage", 75.5); err != nil {
		log.Error().Msgf("Failed to update metric cpu_usage: %v", err)
	}

	if err := storage.UpdateMetric(ctx, models.CounterType, "requests_total", int64(100)); err != nil {
		log.Error().Msgf("Failed to update metric requests_total: %v", err)
	}

	metricUsecase := srvUsecase.NewMetricUsecase(storage, storage, storage)
	router := router.NewRouter(server.NewServer(metricUsecase, nil))

	tests := []struct {
		name       string
		url        string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Get Existing Gauge",
			url:        "/value/gauge/cpu_usage",
			wantStatus: http.StatusOK,
			wantBody:   "75.5",
		},
		{
			name:       "Get Existing Counter",
			url:        "/value/counter/requests_total",
			wantStatus: http.StatusOK,
			wantBody:   "100",
		},
		{
			name:       "Get Non-Existing Metric",
			url:        "/value/gauge/non_existent",
			wantStatus: http.StatusNotFound,
			wantBody:   "Metric non_existent was not found\n",
		},
		{
			name:       "Get Unknown Metric Type",
			url:        "/value/invalid_type/some_metric",
			wantStatus: http.StatusNotFound,
			wantBody:   "Metric some_metric was not found\n",
		},
		{
			name:       "GET with missing name",
			url:        "/value/gauge/",
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			body, _ := io.ReadAll(rr.Body)
			assert.Equal(t, tt.wantBody, string(body))
		})
	}
}
