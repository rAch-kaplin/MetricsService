// The agent is a service that collects metrics from the system and sends them to the server.
// It uses a memory storage to store metrics and a worker pool to send metrics to the server.
//
// # Command-line flags
// -a, --a string   endpoint HTTP-server addr (default "localhost:8080")
// -k, --k string   key for hash (default "")
// -l, --l int      rate limit (default 10)
// -p, --p int      PollInterval value (default 2)
// -r, --r int      PollInterval value (default 10)
//
// Author rAch-kaplin
// Version 1.0.0
// Since 2025-07-29

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	agCfg "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config/agent"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/agent"
	repo "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/repository"
	auc "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/agent"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	workerpool "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/worker-pool"
	pb "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/grpc-metrics"
)

var (
	httpAddress    string
	grpcAddress    string
	pollInterval   int
	reportInterval int
	rateLimit      int
	key            string
	opts           *agCfg.Options
)

var rootCmd = &cobra.Command{
	Use:     "agent",
	Short:   "MetricService",
	Long:    "MetricService",
	Args:    cobra.NoArgs,
	PreRunE: preRunE,
	RunE:    runE,
}

func init() {
	rootCmd.Flags().StringVarP(&httpAddress, "a", "a", agCfg.DefaultHTTPAddress, "endpoint HTTP-server addr")
	rootCmd.Flags().StringVarP(&grpcAddress, "g", "g", agCfg.DefaultGRPCAddress, "endpoint GRPC-server addr")
	rootCmd.Flags().IntVarP(&pollInterval, "p", "p", agCfg.DefaultPollInterval, "PollInterval value")
	rootCmd.Flags().IntVarP(&reportInterval, "r", "r", agCfg.DefaultReportInterval, "PollInterval value")
	rootCmd.Flags().StringVarP(&key, "k", "k", agCfg.DefaultKey, "key for hash")
	rootCmd.Flags().IntVarP(&rateLimit, "l", "l", agCfg.DefaultRateLimit, "rate limit")
}

func preRunE(cmd *cobra.Command, args []string) error {
	var err error
	opts, err = agCfg.ParseOptionsFromCmdAndEnvs(cmd, &agCfg.Options{
		HTTPAddress:    httpAddress,
		GRPCAddress:    grpcAddress,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Key:            key,
		RateLimit:      rateLimit})

	opts = agCfg.NewAgentOptions(
		agCfg.WithAddress(opts.HTTPAddress),
		agCfg.WithGRPCAddress(opts.GRPCAddress),
		agCfg.WithPollInterval(opts.PollInterval),
		agCfg.WithReportInterval(opts.ReportInterval),
		agCfg.WithRateLimit(opts.RateLimit),
		agCfg.WithKey(opts.Key),
	)

	return err
}

func runE(cmd *cobra.Command, args []string) error {
	logFile, err := log.InitLogger("logFileAgent.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logger init error: %v\n", err)
		os.Exit(1)
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

	go func() {
		<-stop
		cancel()
	}()

	startAgent(ctx)

	return nil
}

func startAgent(ctx context.Context) {
	metricStorage := repo.NewMemStorage()
	agentUsecase := agent.NewAgent(auc.NewAgentUsecase(metricStorage, metricStorage))

	client := resty.New().
		SetTimeout(5 * time.Second).
		SetBaseURL("http://" + opts.HTTPAddress)

	wp := workerpool.New(opts.RateLimit)

	wp.Start(ctx)
	defer wp.Wait()

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		agent.CollectMetrics(ctx, agentUsecase, opts.PollInterval)
	}()

	go func() {
		defer wg.Done()
		agent.SendMetrics(ctx, agentUsecase, client, wp, opts.ReportInterval, opts.Key)
	}()

	conn, err := grpc.NewClient(opts.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to GRPC server")
	}
	defer func() {
		_ = conn.Close()
	}()

	grpcClient := pb.NewMetricsServiceClient(conn)

	go func() {
		defer wg.Done()
		agent.SendMetricsGRPC(ctx, agentUsecase, grpcClient, wp, opts.ReportInterval)
	}()

	<-ctx.Done()

	wg.Wait()

	log.Info().Msg("Agent stopped gracefully.")
}
