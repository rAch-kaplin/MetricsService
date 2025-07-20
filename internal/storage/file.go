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
	files "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/files"
	"github.com/rs/zerolog/log"
)

type FileStorage struct {
	mutex      sync.RWMutex
	wg         *sync.WaitGroup
	filePath   string
	storage    col.Collector
	SyncRecord bool
}

type FileParams struct {
	FileStoragePath string
	RestoreOnStart  bool
	StoreInterval   int
}

func (fs *FileStorage) save(ctx context.Context) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if err := files.SaveToDB(ctx, fs.storage, fs.filePath); err != nil {
		log.Error().Err(err).Msg("failed to save DB")
	}
}

func NewFileStorage(ctx context.Context, fp *FileParams) (col.Collector, error) {
	fs := &FileStorage{
		wg:         &sync.WaitGroup{},
		filePath:   fp.FileStoragePath,
		storage:    NewMemStorage(),
		SyncRecord: fp.StoreInterval == 0,
	}
	if fp.RestoreOnStart {
		err := files.LoadFromDB(ctx, fs.storage, fp.FileStoragePath)

		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("LoadFromDB error %w", err)
		}
	}

	if fp.StoreInterval > 0 {
		fs.wg.Add(1)

		go func() {
			defer fs.wg.Done()
			ticker := time.NewTicker(time.Duration(fp.StoreInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					fs.save(ctx)
				case <-ctx.Done():
					log.Info().Msg("Shutting down server, saving metrics")
					fs.save(ctx)
					return
				}
			}
		}()

		fs.wg.Wait()
	}

	return fs, nil
}

func (fs *FileStorage) UpdateMetric(ctx context.Context, mType, mName string, mValue any) error {
	if err := fs.storage.UpdateMetric(ctx, mType, mName, mValue); err != nil {
		log.Error().Err(err).Msg("failed update metric from file storage")
		return fmt.Errorf("failed update metric from file storage %w", err)
	}

	if fs.SyncRecord {
		fs.mutex.Lock()
		defer fs.mutex.Unlock()

		if err := files.SaveToDB(ctx, fs.storage, fs.filePath); err != nil {
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

func (fs *FileStorage) GetMetric(ctx context.Context, mType, mName string) (any, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	val, err := fs.storage.GetMetric(ctx, mType, mName)
	if err != nil {
		log.Error().
			Str("type", mType).
			Str("name", mName).
			Msg("can't get valid metric")
		return nil, fmt.Errorf("can't get valid metric: %v", err)
	}

	return val, nil
}

func (fs *FileStorage) GetAllMetrics(ctx context.Context) []mtr.Metric {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	return fs.storage.GetAllMetrics(ctx)
}

func (fs *FileStorage) Ping(_ context.Context) error {
	return nil
}

func (fs *FileStorage) Close() error {
	return nil
}
