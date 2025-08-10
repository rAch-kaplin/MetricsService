// The server is a service that receives metrics from the agent and stores them in memory.
//
// # Command-line flags
// -a, --a string   endpoint HTTP-server addr (default "localhost:8080")
// -i, --i int      store interval in seconds (0 = sync) (default 300)
// -f, --f string   file to store metrics (default "")
// -r, --r bool     restore metrics from file on start (default true)
// -d, --d string   database dsn (default "")
// -k, --k string   key for hash (default "")
// -t, --t string   trusted subnet (default "")
//
// Author rAch-kaplin
// Version 1.0.0
// Since 2025-07-29
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"

	colcfg "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config/collector"
	srvCfg "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config/server"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/router"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/ping"
	srvUsecase "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/server"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

var (
	endPointAddr    string
	storeInterval   int
	fileStoragePath string
	restoreOnStart  bool
	dataBaseDSN     string
	key             string
	trustedSubnet   string
	opts            *srvCfg.Options
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
	rootCmd.Flags().StringVarP(&endPointAddr, "a", "a", srvCfg.DefaultEndpoint, "endpoint HTTP-server addr")
	rootCmd.Flags().IntVarP(&storeInterval, "i", "i", srvCfg.DefaultStoreInterval, "store interval in seconds (0 = sync)")
	rootCmd.Flags().StringVarP(&fileStoragePath, "f", "f", srvCfg.DefaultFileStoragePath, "file to store metrics")
	rootCmd.Flags().BoolVarP(&restoreOnStart, "r", "r", srvCfg.DefaultRestoreOnStart, "restore metrics from file on start")
	rootCmd.Flags().StringVarP(&dataBaseDSN, "d", "d", srvCfg.DefaultDataBaseDSN, "database dsn")
	rootCmd.Flags().StringVarP(&key, "k", "k", srvCfg.DefaultKey, "key for hash")
	rootCmd.Flags().StringVarP(&trustedSubnet, "t", "t", srvCfg.DefaultTrustedSubnet, "trusted subnet")
}

func preRunE(cmd *cobra.Command, args []string) error {
	var err error
	opts, err = srvCfg.ParseOptionsFromCmdAndEnvs(cmd, &srvCfg.Options{
		EndPointAddr:    endPointAddr,
		StoreInterval:   storeInterval,
		FileStoragePath: fileStoragePath,
		RestoreOnStart:  restoreOnStart,
		DataBaseDSN:     dataBaseDSN,
		Key:             key,
		TrustedSubnet:   trustedSubnet,
	})

	opts = srvCfg.NewServerOptions(
		srvCfg.WithAddress(opts.EndPointAddr),
		srvCfg.WithStoreInterval(opts.StoreInterval),
		srvCfg.WithFileStoragePath(opts.FileStoragePath),
		srvCfg.WithRestoreOnStart(opts.RestoreOnStart),
		srvCfg.WithDataBaseDSN(opts.DataBaseDSN),
		srvCfg.WithKey(opts.Key),
		srvCfg.WithTrustedSubnet(opts.TrustedSubnet),
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

	go func() {
		fmt.Println("pprof listening on :6060")
		_ = http.ListenAndServe("localhost:6060", nil)
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

func startServer(ctx context.Context, opts *srvCfg.Options) error {
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

	metricUsecase := srvUsecase.NewMetricUsecase(collector, collector, collector)

	var pingUsecase *ping.PingUsecase
	if pinger, ok := collector.(ping.Pinger); ok {
		pingUsecase = ping.NewPingUsecase(pinger)
	} else {
		pingUsecase = nil
	}

	r := router.NewRouter(server.NewServer(metricUsecase, pingUsecase), opts)

	srv := &http.Server{
		Addr:    opts.EndPointAddr,
		Handler: r,
	}

	srvErrCh := make(chan error, 1)

	go func() {
		log.Info().Msg("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed unexpectedly")
			srvErrCh <- err
		}
	}()

	select {
	case err := <-srvErrCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Failed to gracefully shutdown server")
			return err
		}

		log.Info().Msg("Server gracefully stopped")
	}

	return nil
}
