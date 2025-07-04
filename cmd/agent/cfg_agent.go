package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/caarlos0/env/v6"
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

	if _, _, err := net.SplitHostPort(opts.endPointAddr); err != nil {
		fmt.Errorf("invalid address %s: %w", opts.endPointAddr, err)
		flag.Usage()
		os.Exit(1)
	}

	return opts
}
