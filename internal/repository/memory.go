package repository

import (
	"context"
	"sync"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
)

// MemStorage is a memory storage for metrics with a mutex for thread safety.
// Data storage is organized as a nested map:
// - the first level of keys is the type of metric (for example, "gauge" or "counter")
// - the second level of keys is the name of the metric
// - the value is an object implementing the models.Metric
type MemStorage struct {
	mutex   sync.RWMutex
	storage map[string]map[string]models.Metric
}

// NewMemStorage creates a new memory storage for metrics
func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: map[string]map[string]models.Metric{
			models.GaugeType:   make(map[string]models.Metric),
			models.CounterType: make(map[string]models.Metric),
		},
	}
}

// updateMetric is a internal function to update a metric in the memory storage.
//
// If the metric is not found, a new metric is created.
// If the metric is found, the value is updated.
func updateMetric(ms *MemStorage, mType, mName string, mValue any) error {
	if _, ok := ms.storage[mType]; !ok {
		return models.ErrInvalidMetricsType
	}

	// If the metric is found, the value is updated.
	if metric, ok := ms.storage[mType][mName]; ok {
		return metric.Update(mValue)
	}

	// If the metric is not found, a new metric is created.
	var newMetric models.Metric

	switch mType {
	case models.GaugeType:
		newMetric = models.NewGauge(mName, 0)
	case models.CounterType:
		newMetric = models.NewCounter(mName, 0)
	default:
		return models.ErrInvalidMetricsType
	}

	// If the metric is created, the value is updated.
	if err := newMetric.Update(mValue); err != nil {
		return err
	}
	ms.storage[mType][mName] = newMetric
	return nil
}

// UpdateMetric updates a metric in the memory storage
func (ms *MemStorage) UpdateMetric(_ context.Context, mType, mName string, mValue any) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	return updateMetric(ms, mType, mName, mValue)
}

// UpdateMetricList updates a list of metrics in the memory storage
func (ms *MemStorage) UpdateMetricList(_ context.Context, metrics []models.Metric) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for _, metric := range metrics {
		if err := updateMetric(ms, metric.Type(), metric.Name(), metric.Value()); err != nil {
			return err
		}
	}

	return nil
}

// GetMetric get a metric from the memory storage
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

// GetAllMetrics get all metrics from the memory storage
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

// Close closes the memory storage
func (ms *MemStorage) Close() error {
	return nil
}
