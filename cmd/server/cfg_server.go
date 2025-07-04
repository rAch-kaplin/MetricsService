package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/caarlos0/env/v6"
)

const (
	defaultEndpoint = "localhost:8080"
)

type options struct {
	endPointAddr string
}

type envConfig struct {
	EndPointAddr string `env:"ADDRESS"`
}

func envAndFlagsInit() *options {
	opts := &options{
		endPointAddr: defaultEndpoint,
	}

	flag.StringVar(&opts.endPointAddr, "a", opts.endPointAddr, "endpoint HTTP-server address")
	flag.Parse()

	var cfg envConfig
	err := env.Parse(&cfg)
	if err != nil {
		fmt.Println("environment variable parsing error")
		os.Exit(1)
	}

	if cfg.EndPointAddr != "" {
		opts.endPointAddr = cfg.EndPointAddr
	}

	if _, _, err := net.SplitHostPort(opts.endPointAddr); err != nil {
		fmt.Errorf("invalid address %s: %w", opts.endPointAddr, err)
		flag.Usage()
		os.Exit(1)
	}

	return opts
}
