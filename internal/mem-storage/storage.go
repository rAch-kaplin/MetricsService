package memstorage

import (
	"sync"
	"maps"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
)

type Collector interface {
	GetMetric(mType, mName string) (any, bool)
	GetAllMetrics() map[string]map[string]metrics.Metric
	UpdateMetric(mtr metrics.Metric) error
}

type MemStorage struct {
	mutex    sync.RWMutex
	storage map[string]map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: make(map[string]map[string]metrics.Metric),
	}
}

func (ms *MemStorage) UpdateMetric(metric metrics.Metric) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	mType := metric.Type()
	mName := metric.Name()

	if _, ok := ms.storage[mType]; !ok {
    	ms.storage[mType] = make(map[string]metrics.Metric)
	}

	if oldMetric, ok := ms.storage[mType][mName]; ok {
		err := oldMetric.Update(metric)	
		if err != nil {
			return err
		}
	} else {
		ms.storage[mType][mName] = metric
	}

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

func (ms *MemStorage) GetAllMetrics() map[string]map[string]metrics.Metric  {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	result := make(map[string]map[string]metrics.Metric, len(ms.storage))

	for mType, innerMap := range ms.storage {
		innerCopy := maps.Clone(innerMap)
		result[mType] = innerCopy
	}

	return result
}

