package metrics

import (
	"errors"
)

type Metric interface {
	Value() interface{}
	Name() string
	Type() string
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
