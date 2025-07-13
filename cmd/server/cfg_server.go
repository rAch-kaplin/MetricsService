package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/router"
	storage "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/storage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

var (
	endPointAddr    string
	storeInterval   int
	fileStoragePath string
	restoreOnStart  bool
	dataBaseDSN     string
	opts            *config.Options
)

var rootCmd = &cobra.Command{
	Use:     "server",
	Short:   "MetricService",
	Long:    "MetricService",
	Args:    cobra.NoArgs,
	PreRunE: PreRunE,
	RunE:    RunE,
}

func init() {
	rootCmd.Flags().StringVarP(&endPointAddr, "a", "a", config.DefaultEndpoint, "endpoint HTTP-server addr")
	rootCmd.Flags().IntVarP(&storeInterval, "i", "i", config.DefaultStoreInterval, "store interval in seconds (0 = sync)")
	rootCmd.Flags().StringVarP(&fileStoragePath, "f", "f", config.DefaultFileStoragePath, "file to store metrics")
	rootCmd.Flags().BoolVarP(&restoreOnStart, "r", "r", config.DefaultRestoreOnStart, "restore metrics from file on start")
	rootCmd.Flags().StringVarP(&dataBaseDSN, "d", "d", config.DefaultDataBaseDSN, "database dsn")
}

func PreRunE(cmd *cobra.Command, args []string) error {
	var err error
	opts, err = config.ParseOptionsFromCmd(cmd, endPointAddr, storeInterval, fileStoragePath, restoreOnStart, dataBaseDSN)
	return err
}

func RunE(cmd *cobra.Command, args []string) error {
	logFile, err := log.InitLogger("logFileServer.log")
	if err != nil {
		return fmt.Errorf("logger init error: %w", err)
	}

	defer func() {
		if err := logFile.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close log file")
		}
	}()

	// log.Info().Msgf("DSN: %s", opts.DataBaseDSN)
	// db, err := sql.Open("pgx", opts.DataBaseDSN)
	// if err != nil {
	// 	log.Error().Err(err).Msg("sql.Open error")
	// 	panic(err)
	// }
	// defer func() {
	// 	if err := db.Close(); err != nil {
	// 		log.Error().Err(err).Msg("Failed to db.Close")
	// 	}
	// }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- startServer(ctx, opts)
	}()

	select {
	case sig := <-stop:
		log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()
		err := <-serverErrCh
		return err
	case err := <-serverErrCh:
		return err
	}
}

func startServer(ctx context.Context, opts *config.Options) error {
	log.Info().
		Str("address", opts.EndPointAddr).
		Msg("Server configuration")

	collector, err := ChoiceCollector(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create collector: %w", err)
	}
	r := router.NewRouter(collector, opts, collector.db)

	srv := &http.Server{
		Addr:    opts.EndPointAddr,
		Handler: r,
	}

	go func() {
		log.Info().Msg("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed unexpectedly")
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Failed to gracefully shutdown server")
		return err
	}

	log.Info().Msg("Server gracefully stopped")

	return nil
}

func ChoiceCollector(ctx context.Context, opts *config.Options) (col.Collector, error) {
	var (
		collector col.Collector
		err       error
	)

	switch {
	// case opts.DataBaseDSN != "":
	// 	collector, err = storage.NewPostgreSQL(ctx, opts.DataBaseDSN)
	case opts.FileStoragePath != "":
		collector, err = storage.NewFileStorage(ctx, &storage.FileParams{
			FileStoragePath: opts.FileStoragePath,
			RestoreOnStart:  opts.RestoreOnStart,
			StoreInterval:   opts.StoreInterval})
	default:
		collector = storage.NewMemStorage()
	}

	if err != nil {
		log.Error().Err(err).Msg("failed create storage")
		return nil, err
	}

	return collector, nil
}
