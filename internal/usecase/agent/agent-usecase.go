package agent

import (
	"context"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
)

type AgentUsecase struct {
	Storage col.Collector
}

func NewAgentUsecase(storage col.Collector) *AgentUsecase {
	return &AgentUsecase{
		Storage: storage,
	}
}

func (uc *AgentUsecase) GetAllMetrics(ctx context.Context) ([]models.Metric, error) {
	return uc.Storage.GetAllMetrics(ctx)
}

func (uc *AgentUsecase) GetMetricJSON(ctx context.Context, metric models.Metric) (*serialize.Metric, error) {
	mType := metric.Type()
	mName := metric.Name()

	metricJSON := &serialize.Metric{
		ID:    mName,
		MType: mType,
	}

	switch mType {
	case models.GaugeType:
		val, ok := metric.Value().(float64)
		if !ok {
			log.Error().Str("metric_name", mName).Str("metric_type", mType).
				Msg("Invalid metric value type")

			return nil, models.ErrInvalidValueType
		}

		metricJSON.Value = &val

	case models.CounterType:
		val, ok := metric.Value().(int64)
		if !ok {
			log.Error().Str("metric_name", mName).Str("metric_type", mType).
				Msg("Invalid metric value type")

			return nil, models.ErrInvalidValueType
		}

		metricJSON.Delta = &val
	}

	return metricJSON, nil
}
