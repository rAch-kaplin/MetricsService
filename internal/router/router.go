package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
)

func NewRouter(srv *server.Server, opts *config.Options) http.Handler {
	r := chi.NewRouter()

	r.Use(server.WithLogging)
	r.Use(server.WithGzipCompress)

	if opts.Key != "" {
		r.Use(server.WithHashing([]byte(opts.Key)))
	}

	r.Route("/", func(r chi.Router) {
		r.Get("/", srv.GetAllMetrics())
		r.Route("/update", func(r chi.Router) {

			r.Post("/", srv.UpdateMetricsHandlerJSON())
			r.Post("/{mType}/{mName}/{mValue}", srv.UpdateMetric())
		})

		r.Route("/value", func(r chi.Router) {
			r.Post("/", srv.GetMetricsHandlerJSON())
			r.Get("/{mType}/{mName}", srv.GetMetric())
		})

		r.Route("/ping", func(r chi.Router) {
			r.Get("/", srv.PingHandler())
		})

		r.Route("/updates", func(r chi.Router) {
			r.Post("/", srv.UpdatesMetricsHandlerJSON())
		})
	})

	return r
}
