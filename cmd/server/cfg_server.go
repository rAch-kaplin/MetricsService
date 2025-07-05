package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/cobra"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
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
		logFile, err := log.InitLogger("logFileServer.log")
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

		if _, _, err := net.SplitHostPort(opts.endPointAddr); err != nil {
			return fmt.Errorf("invalid address %s: %w", opts.endPointAddr, err)
		}

		startServer()

		return nil
	},
}

func init() {
	opts.endPointAddr = defaultEndpoint

	rootCmd.Flags().StringVarP(&opts.endPointAddr, "a", "a", opts.endPointAddr, "endpoint HTTP-server addr")
}

func startServer() {
	log.Info().
		Str("address", opts.endPointAddr).
		Msg("Server configuration")

	storage := ms.NewMemStorage()
	r := server.NewRouter(storage)

	if err := http.ListenAndServe(opts.endPointAddr, r); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP-server didn't start: %v", err)
		panic(err)
	}

}
