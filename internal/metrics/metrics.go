package metrics

import (
	"errors"
)

type Metric interface {
	Value() any
	Name() string
	Type() string
	Update(mtr Metric) error
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
