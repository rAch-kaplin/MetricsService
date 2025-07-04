package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memstorage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func main() {
	log.Init(log.DebugLevel, "logFileServer.log")
	defer log.Destroy()

	opts := envAndFlagsInit()
	log.Debug("Server configuration: Address=%s", opts.endPointAddr)

	log.Debug("START SERVER>")
	fmt.Printf("Endpoint: [%s]\n", opts.endPointAddr)

	storage := ms.NewMemStorage()
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetAllMetrics(storage))
		r.Route("/", func(r chi.Router) {
			r.Get("/value/{mType}/{mName}", server.GetMetric(storage))
			r.Post("/update/{mType}/{mName}/{mValue}", server.UpdateMetric(storage))
		})
	})

	if err := http.ListenAndServe(opts.endPointAddr, r); err != nil {
		log.Error("HTTP-server didn't start: %v", err)
		panic(err)
	}
	log.Debug("ListenAndServe returned")

	log.Debug("END SERVER<")
}
