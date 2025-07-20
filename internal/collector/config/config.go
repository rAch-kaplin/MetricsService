package config

import (
	"context"
	"fmt"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/storage"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

type Params struct {
	Ctx  context.Context
	Opts *config.Options
}

func NewCollector(params *Params) (col.Collector, error) {
	var (
		collector col.Collector
		err       error
	)

	switch {
	case params.Opts.DataBaseDSN != "":
		collector, err = storage.NewDatabase(params.Ctx, params.Opts.DataBaseDSN)
		if err != nil {
			return nil, fmt.Errorf("DB connection failed: %w", err)
		}

	case params.Opts.FileStoragePath != "":
		collector, err = storage.NewFileStorage(params.Ctx, &storage.FileParams{
			FileStoragePath: params.Opts.FileStoragePath,
			RestoreOnStart:  params.Opts.RestoreOnStart,
			StoreInterval:   params.Opts.StoreInterval})

		log.Debug().Msg("chose file storage")

	default:
		collector = storage.NewMemStorage()
	}

	if err != nil {
		log.Error().Err(err).Msg("failed create storage")
		return nil, err
	}

	return collector, nil
}
