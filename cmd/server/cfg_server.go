package main

import (
	"fmt"
	"net"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/cobra"
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

var opts = &options{}

var rootCmd = &cobra.Command{
	Use:   "server",
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

		if _, _, err := net.SplitHostPort(opts.endPointAddr); err != nil {
			return fmt.Errorf("invalid address %s: %w", opts.endPointAddr, err)
		}

		return nil
	},
}

func init() {
	opts.endPointAddr = defaultEndpoint

	rootCmd.Flags().StringVarP(&opts.endPointAddr, "a", "a", opts.endPointAddr, "endpoint HTTP-server addr")
}
