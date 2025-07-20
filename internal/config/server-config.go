package config

import (
	"fmt"
	"net"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/cobra"
)

const (
	DefaultEndpoint        = "localhost:8080"
	DefaultStoreInterval   = 300
	DefaultFileStoragePath = ""
	DefaultRestoreOnStart  = true
	DefaultDataBaseDSN     = ""
)

type Options struct {
	EndPointAddr    string
	StoreInterval   int
	FileStoragePath string
	RestoreOnStart  bool
	DataBaseDSN     string
}

type EnvConfig struct {
	EndPointAddr    string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	RestoreOnStart  bool   `env:"RESTORE"`
	DataBaseDSN     string `env:"DATABASE_DSN"`
}

type Option func(*Options)

func NewServerOptions(options ...Option) *Options {
	opts := &Options{
		EndPointAddr:    DefaultEndpoint,
		StoreInterval:   DefaultStoreInterval,
		FileStoragePath: DefaultFileStoragePath,
		RestoreOnStart:  DefaultRestoreOnStart,
		DataBaseDSN:     DefaultDataBaseDSN,
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

func WithStoreInterval(interval int) Option {
	return func(o *Options) {
		o.StoreInterval = interval
	}
}

func WithFileStoragePath(path string) Option {
	return func(o *Options) {
		o.FileStoragePath = path
	}
}

func WithRestoreOnStart(restore bool) Option {
	return func(o *Options) {
		o.RestoreOnStart = restore
	}
}

func WithDataBaseDSN(dsn string) Option {
	return func(o *Options) {
		o.DataBaseDSN = dsn
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

	if cmd.Flags().Changed("d") {
		opts.DataBaseDSN = src.DataBaseDSN
	}
	if cmd.Flags().Changed("a") {
		opts.EndPointAddr = src.EndPointAddr
	}
	if cmd.Flags().Changed("i") {
		if src.StoreInterval < 0 {
			return nil, fmt.Errorf("store interval must be >= 0, got %d", src.StoreInterval)
		}
		opts.StoreInterval = src.StoreInterval
	}
	if cmd.Flags().Changed("f") {
		if src.FileStoragePath == "" {
			return nil, fmt.Errorf("file storage path flag cannot be empty")
		}
		opts.FileStoragePath = src.FileStoragePath
	}
	if cmd.Flags().Changed("r") {
		opts.RestoreOnStart = src.RestoreOnStart
	}

	return &opts, nil
}

func ParseEnvs(cmd *cobra.Command, opts *Options) error {
	var envCfg EnvConfig
	if err := env.Parse(&envCfg); err != nil {
		return fmt.Errorf("failed to parse environment: %w", err)
	}

	if envCfg.DataBaseDSN != "" {
		cleanDSN := strings.Trim(envCfg.DataBaseDSN, "'")
		opts.DataBaseDSN = cleanDSN
	}

	if envCfg.EndPointAddr != "" {
		opts.EndPointAddr = envCfg.EndPointAddr
	}
	if envCfg.StoreInterval > 0 {
		opts.StoreInterval = envCfg.StoreInterval
	}
	if envCfg.FileStoragePath != "" {
		opts.FileStoragePath = envCfg.FileStoragePath
	}
	opts.RestoreOnStart = envCfg.RestoreOnStart

	return nil
}
