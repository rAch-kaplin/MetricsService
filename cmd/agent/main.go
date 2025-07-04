package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/agent"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memstorage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func main() {
	if err := log.Init(log.DebugLevel, "logFileAgent.log"); err != nil {
		fmt.Errorf("Error initializing the log file: %v", err)
	}
	defer log.Destroy()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	log.Debug("START AGENT>")
	storage := ms.NewMemStorage()

	log.Debug("Configuration: endPointAddr=%s, pollInterval=%ds, reportInterval=%ds",
		opts.endPointAddr, opts.pollInterval, opts.reportInterval)

	client := resty.New().
		SetTimeout(5 * time.Second).
		SetBaseURL("http://" + opts.endPointAddr)

	go agent.CollectionLoop(storage, time.Duration(opts.pollInterval)*time.Second)
	go agent.ReportLoop(client, storage, time.Duration(opts.reportInterval)*time.Second)

	log.Debug("END AGENT<")
	select {}
}
