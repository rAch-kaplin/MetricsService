package agent

import (
	"context"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
)

type MetricGetter interface {
	GetMetric(ctx context.Context, mType, mName string) (models.Metric, error)
	GetAllMetrics(ctx context.Context) ([]models.Metric, error)
}

type MetricUpdater interface {
	UpdateMetric(ctx context.Context, mType, mName string, mValue any) error
	UpdateMetricList(ctx context.Context, metrics []models.Metric) error
}
