package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
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

type envConfig struct {
	endPointAddr string `env:"ADDRESS"`
}

func envAndFlagsInit() *options {
	var cfg envConfig
	err := env.Parse(&cfg)
	if err != nil {
		fmt.Println("environment variable parsing error")
		os.Exit(1)
	}

	opts := &options{
		endPointAddr: defaultEndpoint,
	}

	if cfg.endPointAddr != "" {
		opts.endPointAddr = cfg.endPointAddr
	}

	flag.StringVar(&opts.endPointAddr, "a", defaultEndpoint, "endpoint HTTP-server address")
	flag.Parse()

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
	log.Init(log.DebugLevel, "logFile.log")
	defer log.Destroy()

	opts := envAndFlagsInit()
	log.Debug("Server configuration: Address=%s", opts.endPointAddr)

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
