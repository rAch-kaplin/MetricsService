package database

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"

	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func SaveToDB(collector ms.Collector, path string) error {
	allMetrics := collector.GetAllMetrics()

	data := make([]server.Metrics, 0, len(allMetrics))

	for _, metric := range allMetrics {
		var newMetric server.Metrics

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

	bytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return fmt.Errorf("json Marshal Indent err: %w", err)
	}

	return os.WriteFile(path, bytes, 0666)
}

func LoadFromDB(collector ms.Collector, path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't read file %s with DB %w", path, err)
	}

	var data []server.Metrics
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
