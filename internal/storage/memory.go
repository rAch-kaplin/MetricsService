package storage

import (
	"context"
	"sync"

	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
)

type MemStorage struct {
	mutex   sync.RWMutex
	storage map[string]map[string]mtr.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: make(map[string]map[string]mtr.Metric),
	}
}

func (ms *MemStorage) UpdateMetric(_ context.Context, mType, mName string, mValue any) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, ok := ms.storage[mType]; !ok {
		ms.storage[mType] = make(map[string]mtr.Metric)
	}

	if oldMetric, ok := ms.storage[mType][mName]; ok {
		return oldMetric.Update(mValue)
	}

	var newMetric mtr.Metric

	switch mType {
	case mtr.GaugeType:
		value, ok := mValue.(float64)
		if !ok {
			return mtr.ErrInvalidValueType
		}
		newMetric = mtr.NewGauge(mName, value)
	case mtr.CounterType:
		value, ok := mValue.(int64)
		if !ok {
			return mtr.ErrInvalidValueType
		}
		newMetric = mtr.NewCounter(mName, value)
	default:
		return mtr.ErrInvalidMetricsType
	}

	ms.storage[mType][mName] = newMetric

	return nil
}

func (ms *MemStorage) GetMetric(_ context.Context, mType, mName string) (any, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	if metric, ok := ms.storage[mType][mName]; ok {
		return metric.Value(), true
	}

	return nil, false
}

func (ms *MemStorage) GetAllMetrics(_ context.Context) []mtr.Metric {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	total := 0
	for _, innerMap := range ms.storage {
		total += len(innerMap)
	}
	result := make([]mtr.Metric, 0, total)

	for _, innerMap := range ms.storage {
		for _, metric := range innerMap {
			result = append(result, metric)
		}
	}

	return result
}
