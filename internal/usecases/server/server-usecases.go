package server

import (
	"context"
	"fmt"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
)

type MetricUsecase struct {
	getter  MetricGetter
	updater MetricUpdater
	closer  Closer
}

func NewMetricUsecase(g MetricGetter, u MetricUpdater, c Closer) *MetricUsecase {
	return &MetricUsecase{
		getter:  g,
		updater: u,
		closer:  c,
	}
}

func (uc *MetricUsecase) GetMetric(ctx context.Context, mType, mName string) (models.Metric, error) {
	metric, err := uc.getter.GetMetric(ctx, mType, mName)
	if err != nil {
		return nil, fmt.Errorf("metric not found: %w", err)
	}

	return metric, nil
}

func (uc *MetricUsecase) GetAllMetrics(ctx context.Context) ([]models.Metric, error) {
	allMetrics, err := uc.getter.GetAllMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all metrics: %w", err)
	}

	return allMetrics, nil
}

func (uc *MetricUsecase) UpdateMetric(ctx context.Context, mType, mName string, value any) error {
	if err := uc.updater.UpdateMetric(ctx, mType, mName, value); err != nil {
		return fmt.Errorf("failed to update metric: %w", err)
	}

	return nil
}

func (uc *MetricUsecase) UpdateMetricList(ctx context.Context, metrics []models.Metric) error {
	if err := uc.updater.UpdateMetricList(ctx, metrics); err != nil {
		return fmt.Errorf("failed to update metric list: %w", err)
	}

	return nil
}
