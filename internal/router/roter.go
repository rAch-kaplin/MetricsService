package router

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	database "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/data-base"
)

func NewRouter(storage ms.Collector, opts *config.Options, db *sql.DB) http.Handler {
	r := chi.NewRouter()

	r.Use(server.WithLogging)
	r.Use(server.WithGzipCompress)

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetAllMetrics(storage))
		r.Route("/update", func(r chi.Router) {
			if opts.StoreInterval == 0 {
				r.Use(database.WithSaveToDB(storage, opts.FileStoragePath))
			}

			r.Post("/", server.UpdateMetricsHandlerJSON(storage))
			r.Post("/{mType}/{mName}/{mValue}", server.UpdateMetric(storage))
		})

		r.Route("/value", func(r chi.Router) {
			r.Post("/", server.GetMetricsHandlerJSON(storage))
			r.Get("/{mType}/{mName}", server.GetMetric(storage))
		})

		r.Route("/ping", func(r chi.Router) {
			r.Get("/", server.WithDataBase(db, server.PingDataBase))
		})
	})

	return r
}
