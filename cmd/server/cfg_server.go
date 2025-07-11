package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
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
		var envCfg config.EnvConfig
		if err := env.Parse(&envCfg); err != nil {
			return fmt.Errorf("failed to parse environment: %w", err)
		}

		var envOpts []func(*config.Options)

		if envCfg.EndPointAddr != "" {
			envOpts = append(envOpts, config.WithAddress(envCfg.EndPointAddr))
		}
		if envCfg.StoreInterval != 0 {
			envOpts = append(envOpts, config.WithStoreInterval(envCfg.StoreInterval))
		}
		if envCfg.FileStoragePath != "" {
			envOpts = append(envOpts, config.WithFileStoragePath(envCfg.FileStoragePath))
		}
		envOpts = append(envOpts, config.WithRestoreOnStart(envCfg.RestoreOnStart))

		var flagOpts []func(*config.Options)

		if cmd.Flags().Changed("a") {
			flagOpts = append(flagOpts, config.WithAddress(endPointAddr))
		}
		if cmd.Flags().Changed("i") {
			flagOpts = append(flagOpts, config.WithStoreInterval(storeInterval))
		}
		if cmd.Flags().Changed("f") {
			flagOpts = append(flagOpts, config.WithFileStoragePath(fileStoragePath))
		}
		if cmd.Flags().Changed("r") {
			flagOpts = append(flagOpts, config.WithRestoreOnStart(restoreOnStart))
		}

		opts = config.NewOptions(envOpts, flagOpts)

		if _, _, err := net.SplitHostPort(opts.EndPointAddr); err != nil {
			return fmt.Errorf("invalid address %s: %w", opts.EndPointAddr, err)
		}

		return nil
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

func NewRouter(storage ms.Collector, opts *config.Options) http.Handler {
	r := chi.NewRouter()

	r.Use(server.WithLogging)
	r.Use(server.WithGzipCompress)

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetAllMetrics(storage))
		r.Route("/update", func(r chi.Router) {
			if opts.StoreInterval == 0 {
				r.Use(db.WithSaveToDB(storage, opts.FileStoragePath))
			}

			r.Post("/", server.UpdateMetricsHandlerJSON(storage))
			r.Post("/{mType}/{mName}/{mValue}", server.UpdateMetric(storage))
		})

		r.Route("/value", func(r chi.Router) {
			r.Post("/", server.GetMetricsHandlerJSON(storage))
			r.Get("/{mType}/{mName}", server.GetMetric(storage))
		})
	})

	return r
}

func startServer(opts *config.Options) error {
	log.Info().
		Str("address", opts.EndPointAddr).
		Msg("Server configuration")

	storage := ms.NewMemStorage()
	r := NewRouter(storage, opts)

	if opts.RestoreOnStart {
		if err := db.LoadFromDB(storage, opts.FileStoragePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("LoadFromDB error %w", err)
		}
	}

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

	if err := http.ListenAndServe(opts.EndPointAddr, r); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP-server didn't start: %v", err)
		panic(err)
	}

	return nil
}
