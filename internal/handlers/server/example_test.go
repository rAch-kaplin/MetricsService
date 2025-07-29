package server_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	repo "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/repository"
	srvUsecase "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/server"
)

func Example() {
	// Create memory storage for metrics
	storage := repo.NewMemStorage()

	// Initialize metric use case
	metricUsecase := srvUsecase.NewMetricUsecase(storage, storage, storage)

	// Create server with the metric use case
	srv := server.NewServer(metricUsecase, nil)

	// Create an HTTP POST request to update gauge metric "Alloc" to value 200
	req, _ := http.NewRequest("POST", "/update/gauge/Alloc/200", nil)

	// Create a ResponseRecorder to capture the response
	w := httptest.NewRecorder()

	// Create a new router and register the UpdateMetric handler
	router := chi.NewRouter()
	router.Post("/update/{mType}/{mName}/{mValue}", srv.UpdateMetric())

	// Serve the update request
	router.ServeHTTP(w, req)

	// Create an HTTP GET request to get the updated metric value
	reqGet, _ := http.NewRequest("GET", "/value/gauge/Alloc", nil)
	wGet := httptest.NewRecorder()

	// Register the GetMetric handler on the same router
	router.Get("/value/{mType}/{mName}", srv.GetMetric())

	// Serve the get request
	router.ServeHTTP(wGet, reqGet)

	// Read and print the response body after getting the metric
	bodyGet, _ := io.ReadAll(wGet.Body)
	fmt.Println(string(bodyGet))

	// Output:
	// 200
}
