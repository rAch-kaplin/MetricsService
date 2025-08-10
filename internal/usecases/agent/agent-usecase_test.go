package agent_test

import (
	"context"
	"testing"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	agent "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/agent"
	modelsMocks "github.com/rAch-kaplin/mipt-golang-course/MetricsService/test/mocks/models"
	agentMocks "github.com/rAch-kaplin/mipt-golang-course/MetricsService/test/mocks/usecase/agent"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAgentUsecase_GetAllMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetricGetter := agentMocks.NewMockMetricGetter(ctrl)
	mockMetricUpdater := agentMocks.NewMockMetricUpdater(ctrl)

	uc := agent.NewAgentUsecase(mockMetricUpdater, mockMetricGetter)
	ctx := context.Background()

	t.Run("TestAgentUsecase_GetAllMetrics", func(t *testing.T) {
		mockMetric1 := modelsMocks.NewMockMetric(ctrl)
		mockMetric1.EXPECT().Name().Return("Alloc").AnyTimes()
		mockMetric1.EXPECT().Type().Return("gauge").AnyTimes()
		mockMetric1.EXPECT().Value().Return(124.2).AnyTimes()

		mockMetric2 := modelsMocks.NewMockMetric(ctrl)
		mockMetric2.EXPECT().Name().Return("PollCount").AnyTimes()
		mockMetric2.EXPECT().Type().Return("counter").AnyTimes()
		mockMetric2.EXPECT().Value().Return(int64(100)).AnyTimes()

		mockMetric3 := modelsMocks.NewMockMetric(ctrl)
		mockMetric3.EXPECT().Name().Return("GCCPUFraction").AnyTimes()
		mockMetric3.EXPECT().Type().Return("gauge").AnyTimes()
		mockMetric3.EXPECT().Value().Return(51.1).AnyTimes()

		expectedMetrics := []models.Metric{
			mockMetric1,
			mockMetric2,
			mockMetric3,
		}

		mockMetricGetter.EXPECT().GetAllMetrics(ctx).Return(expectedMetrics, nil)

		metrics, err := uc.GetAllMetrics(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedMetrics, metrics)
	})
}

func TestAgentUsecase_GetMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetricGetter := agentMocks.NewMockMetricGetter(ctrl)
	mockMetricUpdater := agentMocks.NewMockMetricUpdater(ctrl)

	ctx := context.Background()
	uc := agent.NewAgentUsecase(mockMetricUpdater, mockMetricGetter)

	t.Run("TestAgentUsecase_GetMetric_gauge", func(t *testing.T) {
		mockMetric := modelsMocks.NewMockMetric(ctrl)
		mockMetric.EXPECT().Name().Return("Alloc").AnyTimes()
		mockMetric.EXPECT().Type().Return("gauge").AnyTimes()
		mockMetric.EXPECT().Value().Return(124.2).AnyTimes()

		mockMetricGetter.EXPECT().GetMetric(ctx, "gauge", "Alloc").Return(mockMetric, nil)

		metric, err := uc.GetMetric(ctx, "gauge", "Alloc")
		assert.NoError(t, err)
		assert.Equal(t, mockMetric, metric)
	})

	t.Run("TestAgentUsecase_GetMetric_counter", func(t *testing.T) {
		mockMetric := modelsMocks.NewMockMetric(ctrl)
		mockMetric.EXPECT().Name().Return("PollCount").AnyTimes()
		mockMetric.EXPECT().Type().Return("counter").AnyTimes()
		mockMetric.EXPECT().Value().Return(int64(100)).AnyTimes()

		mockMetricGetter.EXPECT().GetMetric(ctx, "counter", "PollCount").Return(mockMetric, nil)

		metric, err := uc.GetMetric(ctx, "counter", "PollCount")
		assert.NoError(t, err)
		assert.Equal(t, mockMetric, metric)
	})
}

func TestAgentUsecase_UpdateMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetricGetter := agentMocks.NewMockMetricGetter(ctrl)
	mockMetricUpdater := agentMocks.NewMockMetricUpdater(ctrl)

	ctx := context.Background()
	uc := agent.NewAgentUsecase(mockMetricUpdater, mockMetricGetter)

	t.Run("TestAgentUsecase_UpdateMetric_gauge", func(t *testing.T) {
		mockMetric := modelsMocks.NewMockMetric(ctrl)
		mockMetric.EXPECT().Name().Return("Alloc").AnyTimes()
		mockMetric.EXPECT().Type().Return("gauge").AnyTimes()
		mockMetric.EXPECT().Value().Return(124.2).AnyTimes()

		mockMetricUpdater.EXPECT().UpdateMetric(ctx, "gauge", "Alloc", 124.2).Return(nil)

		err := uc.UpdateMetric(ctx, "gauge", "Alloc", 124.2)
		assert.NoError(t, err)

		mockMetricGetter.EXPECT().GetMetric(ctx, "gauge", "Alloc").Return(mockMetric, nil)
		metric, err := uc.GetMetric(ctx, "gauge", "Alloc")
		assert.NoError(t, err)
		assert.Equal(t, mockMetric, metric)
	})

	t.Run("TestAgentUsecase_UpdateMetric_counter", func(t *testing.T) {
		mockMetric := modelsMocks.NewMockMetric(ctrl)
		mockMetric.EXPECT().Value().Return(int64(100)).AnyTimes()

		mockMetricUpdater.EXPECT().UpdateMetric(ctx, "counter", "PollCount", int64(100)).Return(nil)

		err := uc.UpdateMetric(ctx, "counter", "PollCount", int64(100))
		assert.NoError(t, err)

		mockMetricGetter.EXPECT().GetMetric(ctx, "counter", "PollCount").Return(mockMetric, nil)
		metric, err := uc.GetMetric(ctx, "counter", "PollCount")
		assert.NoError(t, err)
		assert.Equal(t, mockMetric, metric)
	})
}
