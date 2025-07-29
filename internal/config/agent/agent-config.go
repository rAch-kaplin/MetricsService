package agent

import (
	"fmt"
	"net"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/cobra"
)

const (
	DefaultEndpoint       = "localhost:8080"
	DefaultPollInterval   = 2
	DefaultReportInterval = 10
	DefaultKey            = ""
	DefaultRateLimit      = 10
)

type Options struct {
	EndPointAddr   string
	PollInterval   int
	ReportInterval int
	Key            string
	RateLimit      int
}

type EnvConfig struct {
	EndPointAddr   string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

type Option func(*Options)

func NewAgentOptions(options ...Option) *Options {
	opts := &Options{
		EndPointAddr:   DefaultEndpoint,
		PollInterval:   DefaultPollInterval,
		ReportInterval: DefaultReportInterval,
		Key:            DefaultKey,
		RateLimit:      DefaultRateLimit,
	}

	for _, opt := range options {
		opt(opts)
	}

	return opts
}

func WithAddress(addr string) Option {
	return func(o *Options) {
		o.EndPointAddr = addr
	}
}

func WithKey(key string) Option {
	return func(o *Options) {
		o.Key = key
	}
}

func WithPollInterval(pollInterval int) Option {
	return func(o *Options) {
		o.PollInterval = pollInterval
	}
}

func WithReportInterval(reportInterval int) Option {
	return func(o *Options) {
		o.ReportInterval = reportInterval
	}
}

func WithRateLimit(rateLimit int) Option {
	return func(o *Options) {
		o.RateLimit = rateLimit
	}
}

func ParseOptionsFromCmdAndEnvs(cmd *cobra.Command, src *Options) (*Options, error) {
	opts, err := ParseFlags(cmd, src)
	if err != nil {
		return nil, err
	}

	if err := ParseEnvs(cmd, opts); err != nil {
		return nil, err
	}

	if _, _, err := net.SplitHostPort(opts.EndPointAddr); err != nil {
		return nil, fmt.Errorf("invalid address %s: %w", opts.EndPointAddr, err)
	}

	return opts, nil
}

func ParseFlags(cmd *cobra.Command, src *Options) (*Options, error) {
	opts := *src

	if cmd.Flags().Changed("a") {
		opts.EndPointAddr = src.EndPointAddr
	}

	if cmd.Flags().Changed("p") {
		if src.PollInterval > 0 {
			opts.PollInterval = src.PollInterval
		} else {
			return nil, fmt.Errorf("pollInterval need > 0")
		}
	}

	if cmd.Flags().Changed("r") {
		if src.ReportInterval > 0 {
			opts.ReportInterval = src.ReportInterval
		} else {
			return nil, fmt.Errorf("reportInterval need > 0")
		}
	}

	if cmd.Flags().Changed("l") {
		if src.RateLimit > 0 {
			opts.RateLimit = src.RateLimit
		} else {
			return nil, fmt.Errorf("rateLimit need > 0")
		}
	}

	if cmd.Flags().Changed("k") {
		if src.Key != "" {
			opts.Key = src.Key
		}
	}

	return &opts, nil
}

func ParseEnvs(cmd *cobra.Command, opts *Options) error {
	var cfg EnvConfig
	err := env.Parse(&cfg)
	if err != nil {
		return fmt.Errorf("poll and report intervals must be > 0: %v", err)
	}

	if cfg.EndPointAddr != "" {
		opts.EndPointAddr = cfg.EndPointAddr
	}

	if cfg.PollInterval > 0 {
		opts.PollInterval = cfg.PollInterval
	}

	if cfg.ReportInterval > 0 {
		opts.ReportInterval = cfg.ReportInterval
	}

	if cfg.Key != "" {
		opts.Key = cfg.Key
	}

	if cfg.RateLimit > 0 {
		opts.RateLimit = cfg.RateLimit
	} else {
		opts.RateLimit = DefaultRateLimit
	}

	if opts.PollInterval <= 0 || opts.ReportInterval <= 0 {
		return fmt.Errorf("poll and report intervals must be > 0")
	}

	if _, _, err := net.SplitHostPort(opts.EndPointAddr); err != nil {
		return fmt.Errorf("invalid address %s: %w", opts.EndPointAddr, err)
	}

	return nil
}
