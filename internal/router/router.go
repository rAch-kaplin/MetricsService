// Package router provides a router for the MetricsService API.
// It sets up routing paths, middleware, and handlers for the API, using the chi router.
//
// Author rAch-kaplin
// Version 1.0.0
// Since 2025-07-29
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	srvCfg "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config/server"
	rest "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server/REST"
)

// NewRouter creates and returns a new HTTP router configured with all routes and middleware.
//
// Middleware:
// - WithLogging: Logs the request and response.
// - WithGzipCompress: Compresses the response using gzip.
// - WithHashing: Hashes the request body using the key.
// - WithTrustedSubnet: Checks if the request is from a trusted subnet.
//
// Routes:
//
//	[GET]     "/"                          				- returns all metrics
//	[POST]    "/update/"                   				- batch update metrics (JSON payload)
//	[POST]    "/update/{mType}/{mName}/{mValue}" 		- update a single metric by parameters
//	[POST]    "/value/"                   				- get metrics in batch (JSON payload)
//	[GET]     "/value/{mType}/{mName}"   				- get a single metric by type and name
//	[GET]     "/ping/"                   				- health check endpoint
//	[POST]    "/updates/"                				- alternative batch update endpoint (JSON payload)
//
// Returns:
// - http.Handler
func NewRouter(srv *rest.Server, opts *srvCfg.Options) http.Handler {
	r := chi.NewRouter()

	r.Use(rest.WithLogging)
	r.Use(rest.WithGzipCompress)
	r.Use(rest.WithTrustedSubnet(opts.TrustedSubnet))

	if opts.Key != "" {
		r.Use(rest.WithHashing([]byte(opts.Key)))
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
