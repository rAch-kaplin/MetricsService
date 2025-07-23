package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"

	colcfg "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/router"
	srvUsecase "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecase/server"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecase/ping"
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
	PreRunE: preRunE,
	RunE:    runE,
}

func init() {
	rootCmd.Flags().StringVarP(&endPointAddr, "a", "a", config.DefaultEndpoint, "endpoint HTTP-server addr")
	rootCmd.Flags().IntVarP(&storeInterval, "i", "i", config.DefaultStoreInterval, "store interval in seconds (0 = sync)")
	rootCmd.Flags().StringVarP(&fileStoragePath, "f", "f", config.DefaultFileStoragePath, "file to store metrics")
	rootCmd.Flags().BoolVarP(&restoreOnStart, "r", "r", config.DefaultRestoreOnStart, "restore metrics from file on start")
	rootCmd.Flags().StringVarP(&dataBaseDSN, "d", "d", config.DefaultDataBaseDSN, "database dsn")
}

func preRunE(cmd *cobra.Command, args []string) error {
	var err error
	opts, err = config.ParseOptionsFromCmdAndEnvs(cmd, &config.Options{
		EndPointAddr:    endPointAddr,
		StoreInterval:   storeInterval,
		FileStoragePath: fileStoragePath,
		RestoreOnStart:  restoreOnStart,
		DataBaseDSN:     dataBaseDSN})

	opts = config.NewServerOptions(
		config.WithAddress(opts.EndPointAddr),
		config.WithStoreInterval(opts.StoreInterval),
		config.WithFileStoragePath(opts.FileStoragePath),
		config.WithRestoreOnStart(opts.RestoreOnStart),
		config.WithDataBaseDSN(opts.DataBaseDSN),
	)

	return err
}

func runE(cmd *cobra.Command, args []string) error {
	logFile, err := log.InitLogger("logFileServer.log")
	if err != nil {
		return fmt.Errorf("logger init error: %w", err)
	}

	defer func() {
		if err := logFile.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close log file")
		}
	}()

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

	collector, err := colcfg.NewCollector(&colcfg.Params{
		Ctx:  ctx,
		Opts: opts,
	})
	if err != nil {
		return fmt.Errorf("failed to create collector: %w", err)
	}
	defer func() {
		if err := collector.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close collector")
		}
	}()

	metricUsecase := srvUsecase.NewMetricUsecase(collector)

	var pingUsecase *ping.PingUsecase
	if pinger, ok := collector.(ping.Pinger); ok {
		pingUsecase = ping.NewPingUsecase(pinger)
	} else {
		pingUsecase = nil
	}

	r := router.NewRouter(server.NewServer(metricUsecase, pingUsecase))

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
