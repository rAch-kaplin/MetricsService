package collector

import (
	"context"

	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
)

type Collector interface {
	GetMetric(ctx context.Context, mType, mName string) (any, bool)
	GetAllMetrics(ctx context.Context) []mtr.Metric
	UpdateMetric(ctx context.Context, mType, mName string, mValue any) error
}
