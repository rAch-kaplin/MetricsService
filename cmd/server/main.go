package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
  
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memstorage"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/logger"
)

func main() {
	log.Init(log.DebugLevel, "logFile.log")
	defer log.Destroy()

	log.Debug("START SERVER>")

	storage := ms.NewMemStorage()
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetAllMetrics(storage))
		r.Route("/", func(r chi.Router) {
			r.Get("/value/{mType}/{mName}", server.GetMetric(storage))
			r.Post("/update/{mType}/{mName}/{mValue}", server.UpdateMetric(storage))
		})
	})

	if err := http.ListenAndServe(`:8080`, r); err != nil {
		panic(err)
	}
	log.Debug("END SERVER<")
}
