package agent

import (
	"context"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
)

type AgentUsecase struct {
	updater MetricUpdater
	getter  MetricGetter
}

func NewAgentUsecase(u MetricUpdater, g MetricGetter) *AgentUsecase {
	return &AgentUsecase{
		updater: u,
		getter:  g,
	}
}

func (uc *AgentUsecase) GetAllMetrics(ctx context.Context) ([]models.Metric, error) {
	return uc.getter.GetAllMetrics(ctx)
}

func (uc *AgentUsecase) GetMetric(ctx context.Context, mType, mName string) (models.Metric, error) {
	return uc.getter.GetMetric(ctx, mType, mName)
}
