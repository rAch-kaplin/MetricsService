package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memstorage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func main() {
	if err := log.Init(log.DebugLevel, "logFileServer.log"); err != nil {
		fmt.Errorf("Error initializing the log file: %v", err)
	}
	defer log.Destroy()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	log.Debug("Server configuration: Address=%s", opts.endPointAddr)

	log.Debug("START SERVER>")
	fmt.Printf("Endpoint: [%s]\n", opts.endPointAddr)

	storage := ms.NewMemStorage()
	r := server.NewRouter(storage)

	if err := http.ListenAndServe(opts.endPointAddr, r); err != nil {
		log.Error("HTTP-server didn't start: %v", err)
		panic(err)
	}
	log.Debug("ListenAndServe returned")

	log.Debug("END SERVER<")
}
