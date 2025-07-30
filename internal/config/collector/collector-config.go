// Package collector provides a configuration for the collector.
//
// The storage is chosen in the following order:
//   1. Database storage – used if DataBaseDSN is set.
//   2. File storage – used if FileStoragePath is set.
//   3. In-memory storage – used by default if nothing else is configured.
//
// NewCollector reads options from Params and returns the correct storage.
//
// Author rAch-kaplin
// Version 1.0.0
// Since 2025-07-29

package collector

import (
	"context"
	"fmt"

	srvCfg "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/config/server"
	repo "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/repository"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/server"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

type Params struct {
	Ctx  context.Context
	Opts *srvCfg.Options
}

func NewCollector(params *Params) (server.Collector, error) {
	var (
		collector server.Collector
		err       error
	)

	switch {
	case params.Opts.DataBaseDSN != "":
		collector, err = repo.NewDatabase(params.Ctx, params.Opts.DataBaseDSN)
		if err != nil {
			return nil, fmt.Errorf("DB connection failed: %w", err)
		}

	case params.Opts.FileStoragePath != "":
		collector, err = repo.NewFileStorage(params.Ctx, &repo.FileParams{
			FileStoragePath: params.Opts.FileStoragePath,
			RestoreOnStart:  params.Opts.RestoreOnStart,
			StoreInterval:   params.Opts.StoreInterval})

		log.Debug().Msg("chose file storage")

	default:
		collector = repo.NewMemStorage()
	}

	if err != nil {
		log.Error().Err(err).Msg("failed create storage")
		return nil, err
	}

	return collector, nil
}
