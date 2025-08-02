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
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	colcfg "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config/collector"
	srvCfg "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config/server"
	rest "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server/REST"
	gRPC "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server/gRPC"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/router"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/ping"
	srvUsecase "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/server"
	pb "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/grpc-metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

// Variables for the server configuration
var (
	httpAddress     string
	grpcAddress     string
	storeInterval   int
	fileStoragePath string
	restoreOnStart  bool
	dataBaseDSN     string
	key             string
	trustedSubnet   string
	opts            *srvCfg.Options
)

// Root command for the server
var rootCmd = &cobra.Command{
	Use:     "server",
	Short:   "MetricService",
	Long:    "MetricService",
	Args:    cobra.NoArgs,
	PreRunE: preRunE,
	RunE:    runE,
}

func init() {
	rootCmd.Flags().StringVarP(&httpAddress, "a", "a", srvCfg.DefaultHTTPAddress, "endpoint HTTP-server addr")
	rootCmd.Flags().StringVarP(&grpcAddress, "g", "g", srvCfg.DefaultGRPCAddress, "endpoint GRPC-server addr")
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
		HTTPAddress:     httpAddress,
		GRPCAddress:     grpcAddress,
		StoreInterval:   storeInterval,
		FileStoragePath: fileStoragePath,
		RestoreOnStart:  restoreOnStart,
		DataBaseDSN:     dataBaseDSN,
		Key:             key,
		TrustedSubnet:   trustedSubnet,
	})

	opts = srvCfg.NewServerOptions(
		srvCfg.WithAddress(opts.HTTPAddress),
		srvCfg.WithGRPCAddress(opts.GRPCAddress),
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

	// Create a collector for metrics depending on which type of storage is used (file, database, memory).
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

	// Create a use case for metrics, business logic for metrics.
	metricUsecase := srvUsecase.NewMetricUsecase(collector, collector, collector)

	// Create a use case for ping if the collector implements the Pinger interface.
	var pingUsecase *ping.PingUsecase
	if pinger, ok := collector.(ping.Pinger); ok {
		pingUsecase = ping.NewPingUsecase(pinger)
	} else {
		pingUsecase = nil
	}

	// Create a channel for the shutdown signal.
	stop := make(chan os.Signal, 1)
	// Notify the server about the shutdown signal.
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Create a context for the server.
	g, gCtx := errgroup.WithContext(ctx)

	// Create a goroutine for the server, which will wait for the shutdown signal.
	g.Go(func() error {
		select {
		case <-gCtx.Done():
			return gCtx.Err()
		case sig := <-stop:
			log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
			cancel()
		}

		return nil
	})

	// Create a goroutine for the HTTP server.
	g.Go(func() error {
		return startHTTPServer(gCtx, opts, metricUsecase, pingUsecase)
	})

	// Create a goroutine for the GRPC server.
	g.Go(func() error {
		return startGRPCServer(gCtx, opts, metricUsecase, pingUsecase)
	})

	return g.Wait()
}

func startGRPCServer(ctx context.Context,
	opts *srvCfg.Options,
	metricUsecase *srvUsecase.MetricUsecase,
	pingUsecase *ping.PingUsecase) error {

	log.Info().
		Str("address", opts.GRPCAddress).
		Msg("Server configuration")

	listener, err := net.Listen("tcp", opts.GRPCAddress)
	if err != nil {
		log.Error().Err(err).Msg("Failed to listen")
		return fmt.Errorf("failed to listen: %w", err)
	}

	interceptor := []grpc.UnaryServerInterceptor{
		gRPC.WithLogging,
		gRPC.WithTrustedSubnet(opts.TrustedSubnet),
	}
	if opts.Key != "" {
		interceptor = append(interceptor, gRPC.WithHashing([]byte(opts.Key)))
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor...),
	)
	pb.RegisterMetricsServiceServer(grpcServer, &pb.UnimplementedMetricsServiceServer{})

	grpcErrCh := make(chan error, 1)

	go func() {
		log.Info().Msg("Starting GRPC server...")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal().Err(err).Msg("GRPC server failed unexpectedly")
			grpcErrCh <- err
		}
	}()

	select {
	case err := <-grpcErrCh:
		return err
	case <-ctx.Done():
		grpcServer.GracefulStop()
		log.Info().Msg("GRPC server gracefully stopped")
	}

	return nil
}

func startHTTPServer(ctx context.Context,
	opts *srvCfg.Options,
	metricUsecase *srvUsecase.MetricUsecase,
	pingUsecase *ping.PingUsecase) error {

	log.Info().
		Str("address", opts.HTTPAddress).
		Msg("Server configuration")

	r := router.NewRouter(rest.NewServer(metricUsecase, pingUsecase), opts)

	srv := &http.Server{
		Addr:    opts.HTTPAddress,
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
