package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	server "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/server"
	modelsMocks "github.com/rAch-kaplin/mipt-golang-course/MetricsService/test/mocks/models"
	serverMocks "github.com/rAch-kaplin/mipt-golang-course/MetricsService/test/mocks/server"
	"go.uber.org/mock/gomock"
)

func TestServerUsecase_GetMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetricGetter := serverMocks.NewMockMetricGetter(ctrl)
	mockMetricUpdater := serverMocks.NewMockMetricUpdater(ctrl)

	ctx := context.Background()
	uc := server.NewMetricUsecase(mockMetricGetter, mockMetricUpdater, nil)

	t.Run("TestServerUsecase_GetMetric_gauge", func(t *testing.T) {
		mockMetric := modelsMocks.NewMockMetric(ctrl)
		mockMetric.EXPECT().Name().Return("Alloc").AnyTimes()
		mockMetric.EXPECT().Type().Return("gauge").AnyTimes()
		mockMetric.EXPECT().Value().Return(124.2).AnyTimes()

		mockMetricGetter.EXPECT().GetMetric(ctx, "gauge", "Alloc").Return(mockMetric, nil)

		metric, err := uc.GetMetric(ctx, "gauge", "Alloc")
		assert.NoError(t, err)
		assert.Equal(t, mockMetric, metric)
	})

	t.Run("TestServerUsecase_GetMetric_counter", func(t *testing.T) {
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

func TestServerUsecase_GetAllMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetricGetter := serverMocks.NewMockMetricGetter(ctrl)
	mockMetricUpdater := serverMocks.NewMockMetricUpdater(ctrl)

	ctx := context.Background()
	uc := server.NewMetricUsecase(mockMetricGetter, mockMetricUpdater, nil)

	t.Run("TestServerUsecase_GetAllMetrics", func(t *testing.T) {
		mockMetric1 := modelsMocks.NewMockMetric(ctrl)
		mockMetric1.EXPECT().Name().Return("Alloc").AnyTimes()
		mockMetric1.EXPECT().Type().Return("gauge").AnyTimes()
		mockMetric1.EXPECT().Value().Return(124.2).AnyTimes()

		mockMetric2 := modelsMocks.NewMockMetric(ctrl)
		mockMetric2.EXPECT().Name().Return("PollCount").AnyTimes()
		mockMetric2.EXPECT().Type().Return("counter").AnyTimes()
		mockMetric2.EXPECT().Value().Return(int64(100)).AnyTimes()

		mockMetric3 := modelsMocks.NewMockMetric(ctrl)
		mockMetric3.EXPECT().Name().Return("RandomValue").AnyTimes()
		mockMetric3.EXPECT().Type().Return("gauge").AnyTimes()
		mockMetric3.EXPECT().Value().Return(44.2).AnyTimes()

		expectedMetrics := []models.Metric{mockMetric1, mockMetric2, mockMetric3}

		mockMetricGetter.EXPECT().GetAllMetrics(ctx).Return(expectedMetrics, nil)

		metrics, err := uc.GetAllMetrics(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedMetrics, metrics)
	})
}

func TestServerUsecase_UpdateMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetricGetter := serverMocks.NewMockMetricGetter(ctrl)
	mockMetricUpdater := serverMocks.NewMockMetricUpdater(ctrl)

	ctx := context.Background()
	uc := server.NewMetricUsecase(mockMetricGetter, mockMetricUpdater, nil)

	t.Run("TestServerUsecase_UpdateMetric_gauge", func(t *testing.T) {
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

	t.Run("TestServerUsecase_UpdateMetric_counter", func(t *testing.T) {
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

func TestServerUsecase_UpdateMetricList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetricGetter := serverMocks.NewMockMetricGetter(ctrl)
	mockMetricUpdater := serverMocks.NewMockMetricUpdater(ctrl)

	ctx := context.Background()
	uc := server.NewMetricUsecase(mockMetricGetter, mockMetricUpdater, nil)

	t.Run("TestServerUsecase_UpdateMetricList", func(t *testing.T) {
		mockMetric1 := modelsMocks.NewMockMetric(ctrl)
		mockMetric1.EXPECT().Name().Return("Alloc").AnyTimes()
		mockMetric1.EXPECT().Type().Return("gauge").AnyTimes()
		mockMetric1.EXPECT().Value().Return(124.2).AnyTimes()

		mockMetric2 := modelsMocks.NewMockMetric(ctrl)
		mockMetric2.EXPECT().Name().Return("PollCount").AnyTimes()
		mockMetric2.EXPECT().Type().Return("counter").AnyTimes()
		mockMetric2.EXPECT().Value().Return(int64(100)).AnyTimes()

		mockMetric3 := modelsMocks.NewMockMetric(ctrl)
		mockMetric3.EXPECT().Name().Return("RandomValue").AnyTimes()
		mockMetric3.EXPECT().Type().Return("gauge").AnyTimes()
		mockMetric3.EXPECT().Value().Return(44.2).AnyTimes()

		expectedMetrics := []models.Metric{mockMetric1, mockMetric2, mockMetric3}

		mockMetricUpdater.EXPECT().UpdateMetricList(ctx, expectedMetrics).Return(nil)

		err := uc.UpdateMetricList(ctx, expectedMetrics)
		assert.NoError(t, err)

		mockMetricGetter.EXPECT().GetAllMetrics(ctx).Return(expectedMetrics, nil)
		metrics, err := uc.GetAllMetrics(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedMetrics, metrics)
	})
}
