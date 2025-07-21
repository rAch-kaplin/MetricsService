package collector

import (
	"context"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
)

type Collector interface {
	GetMetric(ctx context.Context, mType, mName string) (any, error)
	GetAllMetrics(ctx context.Context) []models.Metric
	UpdateMetric(ctx context.Context, mType, mName string, mValue any) error

	Ping(ctx context.Context) error
	Close() error
}
