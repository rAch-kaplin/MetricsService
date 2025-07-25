package files

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	server "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/server"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
)

func SaveToDB(ctx context.Context, getter server.MetricGetter, path string) error {
	allMetrics, err := getter.GetAllMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to Get metrics")
	}

	data := make(serialize.MetricsList, 0, len(allMetrics))

	jsonMetrics, err := converter.ConvertToSerialization(allMetrics)
	if err != nil {
		log.Error().Err(err).Msg("failed convert metric")
		return fmt.Errorf("convert metrics error: %w", err)
	}

	if len(jsonMetrics) == 0 {
		log.Warn().Msg("No metrics to save, skipping file write")
		return nil
	}

	bytes, err := json.MarshalIndent(jsonMetrics, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	err = WriteFile(path, bytes)
	if err != nil {
		return fmt.Errorf("write file atomic error: %w", err)
	}

	log.Info().
		Str("path", path).
		Int("metrics_saved", len(data)).
		Msg("Metrics successfully saved")

	return nil
}

func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Error().Err(cerr).Msg("failed to close file")
		}
	}()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	return nil
}

func LoadFromDB(ctx context.Context, updater server.MetricUpdater, path string) error {
	log.Info().
		Str("path", path).
		Msg("Trying to load metrics from file")

	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't read file %s with DB %w", path, err)
	}

	if len(bytes) == 0 {
		log.Warn().Msgf("DB file %s is empty, skipping restore", path)
		return os.ErrNotExist
	}

	var data serialize.MetricsList
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return fmt.Errorf("can't parse json format from DB %w", err)
	}

	metrics, err := converter.ConvertMetrics(data)
	if err != nil {
		return fmt.Errorf("convert metrics error: %w", err)
	}

	for _, metric := range metrics {
		if err := updater.UpdateMetric(ctx, metric.Type(), metric.Name(), metric.Value()); err != nil {
			log.Error().Err(err).Msg("update metric error")
			return fmt.Errorf("update metric error %w", err)
		}
	}

	return nil
}
