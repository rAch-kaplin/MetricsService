package main

import (
	"fmt"
	"net"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/cobra"
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
		var cfg envConfig
		err := env.Parse(&cfg)
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
