package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/router"
	db "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/data-base"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

var (
	endPointAddr    string
	storeInterval   int
	fileStoragePath string
	restoreOnStart  bool
	opts            *config.Options
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "MetricService",
	Long:  "MetricService",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		opts, err = config.ParseOptionsFromCmd(cmd, endPointAddr, storeInterval, fileStoragePath, restoreOnStart)
		return err
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		logFile, err := log.InitLogger("logFileServer.log")
		if err != nil {
			return fmt.Errorf("logger init error: %w", err)
		}

		defer func() {
			if err := logFile.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close log file")
			}
		}()

		if err := startServer(opts); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&endPointAddr, "a", "a", config.DefaultEndpoint, "endpoint HTTP-server addr")
	rootCmd.Flags().IntVarP(&storeInterval, "i", "i", config.DefaultStoreInterval, "store interval in seconds (0 = sync)")
	rootCmd.Flags().StringVarP(&fileStoragePath, "f", "f", config.DefaultFileStoragePath, "file to store metrics")
	rootCmd.Flags().BoolVarP(&restoreOnStart, "r", "r", config.DefaultRestoreOnStart, "restore metrics from file on start")
}

func startServer(opts *config.Options) error {
	log.Info().
		Str("address", opts.EndPointAddr).
		Msg("Server configuration")

	storage := ms.NewMemStorage()
	r := router.NewRouter(storage, opts)

	if opts.RestoreOnStart {
		if err := db.LoadFromDB(storage, opts.FileStoragePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("LoadFromDB error %w", err)
		}
	}

	if opts.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(opts.StoreInterval) * time.Second)
			defer ticker.Stop()

			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

			for {
				select {
				case <-ticker.C:
					if err := db.SaveToDB(storage, opts.FileStoragePath); err != nil {
						log.Error().Err(err).Msg("failed to save DB")
					}
				case <-stop:
					log.Info().Msg("Shutting down server, saving metrics")
					if err := db.SaveToDB(storage, opts.FileStoragePath); err != nil {
						log.Error().Err(err).Msg("Failed to save metrics during shutdown")
					}
					os.Exit(0)
				}
			}
		}()
	}

	log.Info().Msg("Starting HTTP server...")
	if err := http.ListenAndServe(opts.EndPointAddr, r); err != nil {
		log.Error().Err(err).Msg("HTTP server failed to start")
		fmt.Fprintf(os.Stderr, "HTTP-server didn't start: %v", err)
		panic(err)
	}

	return nil
}
