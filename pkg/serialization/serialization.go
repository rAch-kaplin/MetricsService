package serialization

import (
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
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
	case models.GaugeType:
		val, ok := value.(float64)
		if !ok {
			log.Error().Msg("not value type of value")
			return models.ErrInvalidValueType
		}

		mtr.Value = &val
		return nil

	case models.CounterType:
		val, ok := value.(int64)
		if !ok {
			log.Error().Msg("not value type of value")
			return models.ErrInvalidValueType
		}

		mtr.Delta = &val
		return nil

	default:
		return models.ErrInvalidMetricsType
	}
}

func (mtr *Metric) GetValue() (any, error) {
	switch mtr.MType {
	case models.GaugeType:
		if mtr.Value == nil {
			return nil, models.ErrMetricsNotFound
		}

		return *mtr.Value, nil

	case models.CounterType:
		if mtr.Delta == nil {
			return nil, models.ErrMetricsNotFound
		}

		return *mtr.Delta, nil

	default:
		return nil, models.ErrInvalidMetricsType
	}
}
