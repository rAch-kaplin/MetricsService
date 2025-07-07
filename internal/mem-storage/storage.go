package memstorage

import (
	"maps"
	"sync"

	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
)

type Collector interface {
	GetMetric(mType, mName string) (any, bool)
	GetAllMetrics() map[string]map[string]mtr.Metric
	UpdateMetric(mType, mName string, mValue any) error
}

type MemStorage struct {
	mutex   sync.RWMutex
	storage map[string]map[string]mtr.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: make(map[string]map[string]mtr.Metric),
	}
}

func (ms *MemStorage) UpdateMetric(mType, mName string, mValue any) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, ok := ms.storage[mType]; !ok {
		ms.storage[mType] = make(map[string]mtr.Metric)
	}

	if oldMetric, ok := ms.storage[mType][mName]; ok {
		return oldMetric.Update(mType, mName, mValue)
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

func (ms *MemStorage) GetMetric(mType, mName string) (any, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	if metric, ok := ms.storage[mType][mName]; ok {
		return metric.Value(), true
	}

	return nil, false
}

func (ms *MemStorage) GetAllMetrics() map[string]map[string]mtr.Metric {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	result := make(map[string]map[string]mtr.Metric, len(ms.storage))

	for mType, innerMap := range ms.storage {
		innerCopy := maps.Clone(innerMap)
		result[mType] = innerCopy
	}

	return result
}
