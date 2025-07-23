package repository

import (
	"context"
	"sync"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
)

type MemStorage struct {
	mutex   sync.RWMutex
	storage map[string]map[string]models.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: map[string]map[string]models.Metric{
			models.GaugeType:   make(map[string]models.Metric),
			models.CounterType: make(map[string]models.Metric),
		},
	}
}

func (ms *MemStorage) UpdateMetric(_ context.Context, mType, mName string, mValue any) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	if mName == "" {
		return models.ErrMetricsNotFound
	}

	if _, ok := ms.storage[mType]; !ok {
		return models.ErrInvalidMetricsType
	}

	if metric, ok := ms.storage[mType][mName]; ok {
		return metric.Update(mValue)
	}

	var newMetric models.Metric

	switch mType {
	case models.GaugeType:
		newMetric = models.NewGauge(mName, 0)
	case models.CounterType:
		newMetric = models.NewCounter(mName, 0)
	default:
		return models.ErrInvalidMetricsType
	}

	if err := newMetric.Update(mValue); err != nil {
		return err
	}
	ms.storage[mType][mName] = newMetric
	return nil
}

func (ms *MemStorage) GetMetric(_ context.Context, mType, mName string) (models.Metric, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	typedMetric, ok := ms.storage[mType]
	if !ok {
		return nil, models.ErrInvalidMetricsType
	}

	metric, ok := typedMetric[mName]
	if !ok {
		return nil, models.ErrMetricsNotFound
	}

	return metric, nil
}

func (ms *MemStorage) GetAllMetrics(_ context.Context) ([]models.Metric, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	total := 0
	for _, innerMap := range ms.storage {
		total += len(innerMap)
	}
	result := make([]models.Metric, 0, total)

	for _, innerMap := range ms.storage {
		for _, metric := range innerMap {
			result = append(result, metric)
		}
	}

	return result, nil
}

func (ms *MemStorage) Close() error {
	return nil
}
