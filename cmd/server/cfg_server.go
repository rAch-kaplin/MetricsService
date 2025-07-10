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

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	db "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/data-base"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

const (
	defaultEndpoint        = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "/temp/metrics-db.json"
	defaultRestoreOnStart  = true
)

type options struct {
	endPointAddr    string
	storeInterval   int
	fileStoragePath string
	restoreOnStart  bool
}

type envConfig struct {
	EndPointAddr    string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	RestoreOnStart  bool   `env:"RESTORE"`
}

var opts = &options{
	endPointAddr:    defaultEndpoint,
	storeInterval:   defaultStoreInterval,
	fileStoragePath: defaultFileStoragePath,
	restoreOnStart:  defaultRestoreOnStart,
}

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "MetricService",
	Long:  "MetricService",
	Args:  cobra.NoArgs,

	PreRunE: func(cmd *cobra.Command, args []string) error {
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

		if cfg.FileStoragePath != "" {
			opts.fileStoragePath = cfg.FileStoragePath
		}

		if cfg.StoreInterval != 0 {
			opts.storeInterval = cfg.StoreInterval
		}

		opts.restoreOnStart = cfg.RestoreOnStart

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
	rootCmd.Flags().StringVarP(&opts.endPointAddr, "a", "a", defaultEndpoint, "endpoint HTTP-server addr")
	rootCmd.Flags().IntVarP(&opts.storeInterval, "i", "i", defaultStoreInterval, "store interval in seconds (0 = sync)")
	rootCmd.Flags().StringVarP(&opts.fileStoragePath, "f", "f", defaultFileStoragePath, "file to store metrics")
	rootCmd.Flags().BoolVarP(&opts.restoreOnStart, "r", "r", defaultRestoreOnStart, "restore metrics from file on start")
}

func NewRouter(storage ms.Collector, opts *options) http.Handler {
	r := chi.NewRouter()

	r.Use(server.WithLogging)
	r.Use(server.WithGzipCompress)

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetAllMetrics(storage))
		r.Route("/update", func(r chi.Router) {
			if opts.storeInterval == 0 {
				r.Use(db.WithSaveToDB(storage, opts.fileStoragePath))
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

func startServer(opts *options) error {
	log.Info().
		Str("address", opts.endPointAddr).
		Msg("Server configuration")

	storage := ms.NewMemStorage()
	r := NewRouter(storage, opts)

	if opts.restoreOnStart {
		if err := db.LoadFromDB(storage, opts.fileStoragePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("LoadFromDB error %w", err)
		}
	}

	go func() {
		ticker := time.NewTicker(time.Duration(opts.storeInterval) * time.Second)
		defer ticker.Stop()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

		for {
			select {
			case <-ticker.C:
				if err := db.SaveToDB(storage, opts.fileStoragePath); err != nil {
					log.Error().Err(err).Msg("failed to save DB")
				}
			case <-stop:
				log.Info().Msg("Shutting down server, saving metrics")
				if err := db.SaveToDB(storage, opts.fileStoragePath); err != nil {
					log.Error().Err(err).Msg("Failed to save metrics during shutdown")
				}
				os.Exit(0)
			}
		}
	}()

	if err := http.ListenAndServe(opts.endPointAddr, r); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP-server didn't start: %v", err)
		panic(err)
	}

	return nil
}
