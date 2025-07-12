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
	DefaultFileStoragePath = "/temp/metrics-db.json"
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

func NewServerOptions(envOpts, flagOpts []func(*Options)) *Options {
	opts := &Options{
		EndPointAddr:    DefaultEndpoint,
		StoreInterval:   DefaultStoreInterval,
		FileStoragePath: DefaultFileStoragePath,
		RestoreOnStart:  DefaultRestoreOnStart,
	}

	for _, opt := range flagOpts {
		opt(opts)
	}

	for _, opt := range envOpts {
		opt(opts)
	}

	return opts
}

func WithAddress(addr string) func(*Options) {
	return func(o *Options) {
		o.EndPointAddr = addr
	}
}

func WithStoreInterval(interval int) func(*Options) {
	return func(o *Options) {
		o.StoreInterval = interval
	}
}

func WithFileStoragePath(path string) func(*Options) {
	return func(o *Options) {
		o.FileStoragePath = path
	}
}

func WithRestoreOnStart(restore bool) func(*Options) {
	return func(o *Options) {
		o.RestoreOnStart = restore
	}
}

func WithDataBaseDSN(dataBaseDSN string) func(*Options) {
	return func(o *Options) {
		o.DataBaseDSN = dataBaseDSN
	}
}

func ParseOptionsFromCmd(cmd *cobra.Command, endPointAddr string, storeInterval int, fileStoragePath string,
						restoreOnStart bool, dataBaseDSN string) (*Options, error) {
	var envCfg EnvConfig
	if err := env.Parse(&envCfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment: %w", err)
	}

	var envOpts []func(*Options)
	if envCfg.DataBaseDSN != "" {
		cleanDSN := strings.Trim(envCfg.DataBaseDSN, "'")
		envOpts = append(envOpts, WithDataBaseDSN(cleanDSN))
	}
	if envCfg.EndPointAddr != "" {
		envOpts = append(envOpts, WithAddress(envCfg.EndPointAddr))
	}
	if envCfg.StoreInterval < 0 {
		return nil, fmt.Errorf("store interval must be >= 0, got %d", envCfg.StoreInterval)
	} else {
		envOpts = append(envOpts, WithStoreInterval(envCfg.StoreInterval))
	}
	if envCfg.FileStoragePath != "" {
		envOpts = append(envOpts, WithFileStoragePath(envCfg.FileStoragePath))
	}
	envOpts = append(envOpts, WithRestoreOnStart(envCfg.RestoreOnStart))

	var flagOpts []func(*Options)
	if cmd.Flags().Changed("d") {
		flagOpts = append(flagOpts, WithDataBaseDSN(dataBaseDSN))
	}
	if cmd.Flags().Changed("a") {
		flagOpts = append(flagOpts, WithAddress(endPointAddr))
	}

	if cmd.Flags().Changed("i") {
		if storeInterval < 0 {
			return nil, fmt.Errorf("store interval flag must be >= 0, got %d", storeInterval)
		}
		flagOpts = append(flagOpts, WithStoreInterval(storeInterval))
	}

	if cmd.Flags().Changed("f") {
		if fileStoragePath == "" {
			return nil, fmt.Errorf("file storage path flag cannot be empty")
		}
		flagOpts = append(flagOpts, WithFileStoragePath(fileStoragePath))
	}

	if cmd.Flags().Changed("r") {
		flagOpts = append(flagOpts, WithRestoreOnStart(restoreOnStart))
	}

	opts := NewServerOptions(envOpts, flagOpts)

	if _, _, err := net.SplitHostPort(opts.EndPointAddr); err != nil {
		return nil, fmt.Errorf("invalid address %s: %w", opts.EndPointAddr, err)
	}

	return opts, nil
}
