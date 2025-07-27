package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/agent"
	repo "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/repository"
	auc "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/agent"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	workerpool "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/worker-pool"
)

const (
	defaultEndpoint       = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultKey            = "key"
)

type options struct {
	endPointAddr   string
	pollInterval   int
	reportInterval int
	key            string
}

type envConfig struct {
	EndPointAddr   string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
}

var opts = &options{}

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "MetricService",
	Long:  "MetricService",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		var cfg envConfig
		err = env.Parse(&cfg)
		if err != nil {
			return fmt.Errorf("poll and report intervals must be > 0")
		}

		if cfg.EndPointAddr != "" {
			opts.endPointAddr = cfg.EndPointAddr
		}
		if cfg.PollInterval > 0 {
			opts.pollInterval = cfg.PollInterval
		}
		if cfg.ReportInterval > 0 {
			opts.reportInterval = cfg.ReportInterval
		}

		if cfg.Key != "" {
			opts.key = cfg.Key
		}

		if opts.pollInterval <= 0 || opts.reportInterval <= 0 {
			return fmt.Errorf("poll and report intervals must be > 0")
		}

		if _, _, err := net.SplitHostPort(opts.endPointAddr); err != nil {
			return fmt.Errorf("invalid address %s: %w", opts.endPointAddr, err)
		}

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
	},
}

func init() {
	opts.endPointAddr = defaultEndpoint
	opts.pollInterval = defaultPollInterval
	opts.reportInterval = defaultReportInterval
	opts.key = defaultKey

	rootCmd.Flags().StringVarP(&opts.endPointAddr, "a", "a", opts.endPointAddr, "endpoint HTTP-server addr")
	rootCmd.Flags().IntVarP(&opts.pollInterval, "p", "p", opts.pollInterval, "PollInterval value")
	rootCmd.Flags().IntVarP(&opts.reportInterval, "r", "r", opts.reportInterval, "PollInterval value")
	rootCmd.Flags().StringVarP(&opts.key, "k", "k", opts.key, "key for hash")
}

func startAgent(ctx context.Context) {
	metricStorage := repo.NewMemStorage()
	agentUsecase := agent.NewAgent(auc.NewAgentUsecase(metricStorage, metricStorage))

	client := resty.New().
		SetTimeout(5 * time.Second).
		SetBaseURL("http://" + opts.endPointAddr)

	wp := workerpool.New(3)

	wp.Start(ctx)
	defer wp.Wait()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		agent.CollectMetrics(ctx, metricStorage, opts.pollInterval)
	}()

	go func() {
		defer wg.Done()
		agent.SendMetrics(ctx, agentUsecase, client, wp, opts.reportInterval, opts.key)
	}()

	<-ctx.Done()

	wg.Wait()

	log.Info().Msg("Agent stopped gracefully.")
}
