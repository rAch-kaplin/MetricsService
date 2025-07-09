package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/agent"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
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

var opts = &options{}

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "MetricService",
	Long:  "MetricService",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logFile, err := log.InitLogger("logFileAgent.log")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Logger init error: %v\n", err)
			os.Exit(1)
		}

		defer func() {
			if err := logFile.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close log file")
			}
		}()

		var cfg envConfig
		err = env.Parse(&cfg)
		if err != nil {
			return fmt.Errorf("poll and report intervals must be > 0")
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
			return fmt.Errorf("poll and report intervals must be > 0")
		}

		if _, _, err := net.SplitHostPort(opts.endPointAddr); err != nil {
			return fmt.Errorf("invalid address %s: %w", opts.endPointAddr, err)
		}

		startAgent()

		return nil
	},
}

func init() {
	opts.endPointAddr = defaultEndpoint
	opts.pollInterval = defaultPollInterval
	opts.reportInterval = defaultReportInterval

	rootCmd.Flags().StringVarP(&opts.endPointAddr, "a", "a", opts.endPointAddr, "endpoint HTTP-server addr")
	rootCmd.Flags().IntVarP(&opts.pollInterval, "p", "p", opts.pollInterval, "PollInterval value")
	rootCmd.Flags().IntVarP(&opts.reportInterval, "r", "r", opts.reportInterval, "PollInterval value")
}

func startAgent() {
	storage := ms.NewMemStorage()

	client := resty.New().
		SetTimeout(5 * time.Second).
		SetBaseURL("http://" + opts.endPointAddr)

	log.Info().Msg("Starting collection and reporting loops")

	go agent.CollectionLoop(storage, time.Duration(opts.pollInterval)*time.Second)
	go agent.ReportLoop(client, storage, time.Duration(opts.reportInterval)*time.Second)

	select {}
}
