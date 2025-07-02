package memstorage

import (
	"sync"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
)

type Collector interface {
	GetMetric(mType, mName string) (interface{}, bool)
	GetAllMetrics() (map[string]float64, map[string]int64)
	UpdateMetric(mtr metrics.Metric) error
}


type MemStorage struct {
	mutex    sync.RWMutex
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateMetric(metric metrics.Metric) error {
	switch metric.Type() {
	case metrics.CounterType:
		{
			val, ok := metric.Value().(int64)
			if !ok {
				return metrics.ErrInvalidValueType
			}

			ms.UpdateCounter(metric, val)
			return nil
		}
	case metrics.GaugeType:
		{
			val, ok := metric.Value().(float64)
			if !ok {
				return metrics.ErrInvalidValueType
			}

			ms.UpdateGauge(metric, val)
			return nil
		}
	default:
		return metrics.ErrInvalidMetricsType
	}
}

func (ms *MemStorage) UpdateGauge(metric metrics.Metric, value float64) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.Gauges[metric.Name()] = value
}

func (ms *MemStorage) UpdateCounter(metric metrics.Metric, value int64) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.Counters[metric.Name()] += value
}

func (ms *MemStorage) GetMetric(mType, mName string) (interface{}, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	switch mType {
	case metrics.GaugeType:
		{
			if val, ok := ms.Gauges[mName]; ok {
				return val, true
			}
		}
	case metrics.CounterType:
		{
			if val, ok := ms.Counters[mName]; ok {
				return val, true
			}
		}
	}

	return nil, false
}

func (ms *MemStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	gaugesCopy := make(map[string]float64, len(ms.Gauges))
	for name, value := range ms.Gauges {
		gaugesCopy[name] = value
	}

	countersCopy := make(map[string]int64, len(ms.Counters))
	for name, value := range ms.Counters {
		countersCopy[name] = value
	}

	return gaugesCopy, countersCopy
}
