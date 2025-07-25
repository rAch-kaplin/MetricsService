package metrics

import (
	"errors"
)
type Metric interface {
	Value() any
	Name() string
	Type() string
	Update(mValue any) error
}

//easyjson:json
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
type MetricTable struct {
	Name  string
	Type  string
	Value string
}

const (
	CounterType = "counter"
	GaugeType   = "gauge"
)

var (
	ErrInvalidMetricsType = errors.New("invalid metrics type")
	ErrInvalidValueType   = errors.New("invalid value type")
	ErrMetricsNotFound    = errors.New("unknown this metric")
)
