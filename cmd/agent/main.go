package main

import (
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/agent"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memstorage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func main() {
	log.Init(log.DebugLevel, "logFileAgent.log")
	defer log.Destroy()

	log.Debug("START AGENT>")
	storage := ms.NewMemStorage()

	opts := envAndFlagsInit()
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
