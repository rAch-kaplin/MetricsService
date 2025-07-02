package main

import (
	"net/http"

	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memStorage"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/logger"
)

func main() {
	log.Init(log.DebugLevel, "logFile.log")
	defer log.Destroy()

	log.Debug("START SERVER>")
	storage := ms.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", server.MainHandle(storage))

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
	log.Debug("END SERVER<")
}
