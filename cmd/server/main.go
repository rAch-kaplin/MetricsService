package main

import (
	"flag"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memstorage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/logger"
)

const (
	defaultEndpoint = "localhost:8080"
)

type options struct {
	endPointAddr string
}

func flagsInit(opts *options) {
	flag.StringVar(&opts.endPointAddr, "a", defaultEndpoint, "endpoint HTTP-server address")
	flag.Parse()
}

func main() {
	log.Init(log.DebugLevel, "logFile.log")
	defer log.Destroy()

	var opts options
	flagsInit(&opts)

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

	if err := http.ListenAndServe(opts.endPointAddr, r); err != nil {
		panic(err)
	}
	log.Debug("END SERVER<")
}
