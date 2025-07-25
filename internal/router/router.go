package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
)

func NewRouter(storage col.Collector, opts *config.Options) http.Handler {
	r := chi.NewRouter()

	r.Use(server.WithLogging)
	r.Use(server.WithGzipCompress)

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetAllMetrics(storage))
		r.Route("/update", func(r chi.Router) {

			r.Post("/", server.UpdateMetricsHandlerJSON(storage))
			r.Post("/{mType}/{mName}/{mValue}", server.UpdateMetric(storage))
		})

		r.Route("/value", func(r chi.Router) {
			r.Post("/", server.GetMetricsHandlerJSON(storage))
			r.Get("/{mType}/{mName}", server.GetMetric(storage))
		})

		r.Route("/ping", func(r chi.Router) {
			r.Get("/", server.PingHandler(storage))
		})
	})

	return r
}
