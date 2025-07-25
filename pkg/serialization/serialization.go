package serialization

import (
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

//easyjson:json
type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

//easyjson:json
type MetricsList []Metric

func (mtr *Metric) SetValue(value any) error {
	switch mtr.MType {
	case metrics.GaugeType:
		val, ok := value.(float64)
		if !ok {
			log.Error().Msg("not value type of value")
			return metrics.ErrInvalidValueType
		}

		mtr.Value = &val
		return nil

	case metrics.CounterType:
		val, ok := value.(int64)
		if !ok {
			log.Error().Msg("not value type of value")
			return metrics.ErrInvalidValueType
		}

		mtr.Delta = &val
		return nil

	default:
		return metrics.ErrInvalidMetricsType
	}
}

func (mtr *Metric) GetValue() (any, error) {
	switch mtr.MType {
	case metrics.GaugeType:
		if mtr.Value == nil {
			return nil, metrics.ErrMetricsNotFound
		}

		return *mtr.Value, nil

	case metrics.CounterType:
		if mtr.Delta == nil {
			return nil, metrics.ErrMetricsNotFound
		}

		return *mtr.Delta, nil

	default:
		return nil, metrics.ErrInvalidMetricsType
	}
}
