package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-resty/resty/v2"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/agent"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memstorage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/logger"
)

const (
	defaultEndpoint       = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

type options struct {
	endPointAddr   string
	pollInterval   int
	reportInterval int
}

type envConfig struct {
	EndPointAddr   string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func envAndFlagsInit() *options {
	opts := &options{
		endPointAddr:   defaultEndpoint,
		pollInterval:   defaultPollInterval,
		reportInterval: defaultReportInterval,
	}

	flag.StringVar(&opts.endPointAddr, "a", opts.endPointAddr, "endpoint HTTP-server addr")
	flag.IntVar(&opts.pollInterval, "p", opts.pollInterval, "PollInterval value")
	flag.IntVar(&opts.reportInterval, "r", opts.reportInterval, "ReportInterval value")

	flag.Parse()

	var cfg envConfig
	err := env.Parse(&cfg)
	if err != nil {
		fmt.Println("environment variables parsing error")
		os.Exit(1)
	}

	if cfg.EndPointAddr != "" {
		opts.endPointAddr = cfg.EndPointAddr
	}
	if cfg.PollInterval > 0 {
		opts.pollInterval = cfg.PollInterval
	}
	if cfg.ReportInterval > 0 {
		opts.reportInterval = cfg.ReportInterval
	}

	if opts.pollInterval <= 0 || opts.reportInterval <= 0 {
		fmt.Println("Error: poll interval and report interval must be > 0")
		flag.Usage()
		os.Exit(1)
	}

	if err := validateEndpoint(opts.endPointAddr); err != nil {
		fmt.Printf("Error in endpoint address: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	return opts
}

func validateEndpoint(addr string) error {
	parts := strings.Split(addr, ":")

	if len(parts) != 2 {
		return fmt.Errorf("address must be in format 'host:port'")
	}

	if parts[0] == "" {
		return fmt.Errorf("host cannot be empty")
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil || port <= 0 {
		fmt.Printf("Error: Port must be >0 number\n")
		flag.Usage()
		os.Exit(1)
	}

	return nil
}

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
