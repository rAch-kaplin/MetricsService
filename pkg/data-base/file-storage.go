package database

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func SaveToDB(collector ms.Collector, path string) error {
	allMetrics := collector.GetAllMetrics()

	data := make([]mtr.Metrics, 0, len(allMetrics))

	for _, metric := range allMetrics {
		var newMetric mtr.Metrics

		newMetric.MType = metric.Type()
		newMetric.ID = metric.Name()

		switch metric.Type() {
		case mtr.GaugeType:
			val, ok := metric.Value().(float64)
			if !ok {
				return fmt.Errorf("invalid type metric")
			}
			newMetric.Value = &val
		case mtr.CounterType:
			val, ok := metric.Value().(int64)
			if !ok {
				return fmt.Errorf("invalid type metric")
			}
			newMetric.Delta = &val
		default:
			log.Error().Msg("unknown metric type")
		}

		data = append(data, newMetric)
	}

	if len(data) == 0 {
		log.Warn().Msg("No metrics to save, skipping file write")
		return nil
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpPath := path + ".tmp"
	file, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close file")
		}
	}()

	if _, err := file.Write(bytes); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}

	log.Info().
		Str("path", path).
		Int("metrics_saved", len(data)).
		Msg("Metrics successfully saved")

	return nil
}

func LoadFromDB(collector ms.Collector, path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't read file %s with DB %w", path, err)
	}

	if len(bytes) == 0 {
		log.Warn().Msgf("DB file %s is empty, skipping restore", path)
		return nil
	}

	var data []mtr.Metrics
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return fmt.Errorf("can't parse json format from DB %w", err)
	}

	for _, metric := range data {
		switch metric.MType {
		case mtr.GaugeType:
			if err := collector.UpdateMetric(metric.MType, metric.ID, *metric.Value); err != nil {
				log.Error().Err(err).Msg("update metric error")
				return fmt.Errorf("update metric error %w", err)
			}
		case mtr.CounterType:
			if err := collector.UpdateMetric(metric.MType, metric.ID, *metric.Delta); err != nil {
				log.Error().Err(err).Msg("update metric error")
				return fmt.Errorf("update metric error %w", err)
			}
		}
	}

	return nil
}

func WithSaveToDB(collector ms.Collector, filePath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req)

			if err := SaveToDB(collector, filePath); err != nil {
				log.Error().Err(err).Msg("Failed to save metrics synchronously")
			}
		})
	}
}
