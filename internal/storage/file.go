package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	database "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/data-base"
	"github.com/rs/zerolog/log"
)

type FileStorage struct {
	mutex      sync.RWMutex
	filePath   string
	storage    col.Collector
	SyncRecord bool
}

type FileParams struct {
	FileStoragePath string
	RestoreOnStart  bool
	StoreInterval   int
}

func NewFileStorage(ctx context.Context, fp *FileParams) (col.Collector, error) {
	fs := &FileStorage{
		filePath:   fp.FileStoragePath,
		storage:    NewMemStorage(),
		SyncRecord: false,
	}
	if fp.RestoreOnStart {
		if err := database.LoadFromDB(ctx, fs.storage, fp.FileStoragePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("LoadFromDB error %w", err)
		}
	}

	if fp.StoreInterval == 0 {
		fs.SyncRecord = true
	}

	if fp.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(fp.StoreInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := database.SaveToDB(ctx, fs.storage, fp.FileStoragePath); err != nil {
						log.Error().Err(err).Msg("failed to save DB")
					}
				case <-ctx.Done():
					log.Info().Msg("Shutting down server, saving metrics")
					if err := database.SaveToDB(ctx, fs.storage, fp.FileStoragePath); err != nil {
						log.Error().Err(err).Msg("Failed to save metrics during shutdown")
					}
					return
				}
			}
		}()
	}

	return fs, nil
}

func (fs *FileStorage) UpdateMetric(ctx context.Context, mType, mName string, mValue any) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if err := fs.storage.UpdateMetric(ctx, mType, mName, mValue); err != nil {
		log.Error().Err(err).Msg("failed update metric from file storage")
		return fmt.Errorf("failed update metric from file storage %w", err)
	}

	if fs.SyncRecord {
		if err := database.SaveToDB(ctx, fs.storage, fs.filePath); err != nil {
			log.Error().Err(err).Msg("failed save storage")
			return fmt.Errorf("failed save storage %w", err)
		}
	}

	log.Info().
		Str("type", mType).
		Str("name", mName).
		Msg("Metric updated successfully")

	return nil
}

func (fs *FileStorage) GetMetric(ctx context.Context, mType, mName string) (any, bool) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	val, ok := fs.storage.GetMetric(ctx, mType, mName)
	if !ok {
		log.Error().
			Str("type", mType).
			Str("name", mName).
			Msg("can't get valid metric")
		return nil, false
	}

	return val, true
}

func (fs *FileStorage) GetAllMetrics(ctx context.Context) []mtr.Metric {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	return fs.storage.GetAllMetrics(ctx)
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (fs *FileStorage) Close() error {
	return nil
}
