package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	db "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/data-base"
)

func NewRouter(storage ms.Collector, opts *config.Options) http.Handler {
	r := chi.NewRouter()

	r.Use(server.WithLogging)
	r.Use(server.WithGzipCompress)

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetAllMetrics(storage))
		r.Route("/update", func(r chi.Router) {
			if opts.StoreInterval == 0 {
				r.Use(db.WithSaveToDB(storage, opts.FileStoragePath))
			}

			r.Post("/", server.UpdateMetricsHandlerJSON(storage))
			r.Post("/{mType}/{mName}/{mValue}", server.UpdateMetric(storage))
		})

		r.Route("/value", func(r chi.Router) {
			r.Post("/", server.GetMetricsHandlerJSON(storage))
			r.Get("/{mType}/{mName}", server.GetMetric(storage))
		})
	})

	return r
}
