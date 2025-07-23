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
		storage: make(map[string]map[string]models.Metric),
	}
}

func (ms *MemStorage) UpdateMetric(_ context.Context, mType, mName string, mValue any) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, ok := ms.storage[mType]; !ok {
		ms.storage[mType] = make(map[string]models.Metric)
	}

	if oldMetric, ok := ms.storage[mType][mName]; ok {
		return oldMetric.Update(mValue)
	}

	var newMetric models.Metric

	switch mType {
	case models.GaugeType:
		value, ok := mValue.(float64)
		if !ok {
			return models.ErrInvalidValueType
		}
		newMetric = models.NewGauge(mName, value)
	case models.CounterType:
		value, ok := mValue.(int64)
		if !ok {
			return models.ErrInvalidValueType
		}
		newMetric = models.NewCounter(mName, value)
	default:
		return models.ErrInvalidMetricsType
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

func (ms *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (ms *MemStorage) Close() error {
	return nil
}
