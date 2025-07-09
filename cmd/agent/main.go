package main

import (
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/agent"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memstorage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/logger"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	log.Init(log.DebugLevel, "logFile.log")
	defer log.Destroy()

	log.Debug("START AGENT>")
	storage := ms.NewMemStorage()

	client := resty.New().SetTimeout(5 * time.Second)

	go agent.CollectionLoop(storage, pollInterval)
	go agent.ReportLoop(client, storage, reportInterval)

	log.Debug("END AGENT<")
	select {}
}
