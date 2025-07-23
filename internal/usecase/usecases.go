package usecase

import (
	"context"
	"fmt"
	"strconv"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/rs/zerolog/log"
)

type MetricUsecase struct {
	repo col.Collector
}

func NewMetricUsecase(repo col.Collector) *MetricUsecase {
	return &MetricUsecase{repo: repo}
}

func (uc *MetricUsecase) GetMetric(ctx context.Context, mType, mName string) (models.Metric, error) {
	metric, err := uc.repo.GetMetric(ctx, mType, mName)
	if err != nil {
		return nil, fmt.Errorf("metric not found: %w", err)
	}

	return metric, nil
}

func (uc *MetricUsecase) GetAllMetrics(ctx context.Context) ([]models.MetricTable, error) {
	allMetrics, err := uc.repo.GetAllMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to Get metrics")
		return nil, fmt.Errorf("failed to get metrics: %v", err)
	}

	var metricsToTable []models.MetricTable

	for _, metric := range allMetrics {
		var valStr string
		mName := metric.Name()
		mType := metric.Type()

		switch mType {
		case models.GaugeType:
			val, ok := metric.Value().(float64)
			if !ok {
				log.Error().Str("metric_name", mName).Str("metric_type", mType).
					Msg("Invalid metric value type")
				continue
			}
			valStr = strconv.FormatFloat(val, 'f', -1, 64)

		case models.CounterType:
			val, ok := metric.Value().(int64)
			if !ok {
				log.Error().Str("metric_name", mName).Str("metric_type", mType).
					Msg("Invalid metric value type")
				continue
			}
			valStr = strconv.FormatInt(val, 10)
		}

		metricsToTable = append(metricsToTable, models.MetricTable{
			Name:  mName,
			Type:  mType,
			Value: valStr,
		})
	}

	return metricsToTable, nil
}

func (uc *MetricUsecase) UpdateMetric(ctx context.Context, mType, mName string, value any) error {
	if err := uc.repo.UpdateMetric(ctx, mType, mName, value); err != nil {
		return fmt.Errorf("failed to update metric: %w", err)
	}

	return nil
}

func (uc *MetricUsecase) UpdateMetricList(ctx context.Context, metrics []models.Metric) error {
	for _, metric := range metrics {
		if err := uc.UpdateMetric(ctx, metric.Type(), metric.Name(), metric.Value()); err != nil {
			return fmt.Errorf("failed to update metric: %w", err)
		}
	}

	return nil
}

func (uc *MetricUsecase) Ping(ctx context.Context) error {
	if err := uc.repo.Ping(ctx); err != nil {
		return fmt.Errorf("failed ping database: %w", err)
	}
	return nil
}
