package agent_test

import (
	"context"
	"testing"

	agent "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/agent"
	repo "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/repository"
	auc "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/agent"
	"github.com/stretchr/testify/assert"
)

func TestAgent_UpdateAllMetrics(t *testing.T) {
	metricStorage := repo.NewMemStorage()
	agent := agent.NewAgent(auc.NewAgentUsecase(metricStorage, metricStorage))

	ctx := context.Background()
	t.Run("Agent_UpdateAllMetrics", func(t *testing.T) {
		agent.UpdateAllMetrics(ctx)

		metrics, err := agent.Usecase.GetAllMetrics(ctx)
		if err != nil {
			t.Errorf("UpdateAllMetrics() error = %v", err)
		}

		names := []string{}
		for _, metric := range metrics {
			names = append(names, metric.Name())
		}

		expected := []string{
			"Alloc", "BuckHashSys", "Frees",
			"GCCPUFraction", "GCSys", "HeapAlloc",
			"HeapIdle", "HeapInuse", "HeapObjects",
			"HeapReleased", "HeapSys", "LastGC",
			"Lookups", "MCacheInuse", "MCacheSys",
			"Mallocs", "NextGC", "NumForcedGC",
			"NumGC", "OtherSys", "PauseTotalNs",
			"StackInuse", "StackSys", "Sys",
			"TotalAlloc", "TotalMemory", "FreeMemory",
			"RandomValue", "CPUutilization1",
			"PollCount",
		}

		for _, exp := range expected {
			assert.Contains(t, names, exp)
		}
	})

}
