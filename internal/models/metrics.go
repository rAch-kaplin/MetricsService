// Package models contains definitions, interfaces, and methods for working with metrics.
// Metrics come in two types: Gauge and Counter.
// The package provides a Metric interface for describing the basic behavior of metrics,
// as well as a MetricTable structure for representing metrics in tabular form.
//
// Author rAch-kaplin
// Version 1.0.0
// Since 2025-07-29
package models

import (
	"errors"
)

// Metric is an interface that describes the basic behavior of metrics.
type Metric interface {
	Value() any
	Name() string
	Type() string
	Update(mValue any) error
}

// MetricTable is a structure for representing metrics in tabular form.
type MetricTable struct {
	Name  string
	Type  string
	Value string
}

// Constants that define the supported metric types.
const (
	// CounterType indicates the type of counter (increasing metric).
	CounterType = "counter"

	// GaugeType indicates the type of gauge (floating-point metric).
	GaugeType = "gauge"
)

// Errors that can be returned by the package.
var (
	ErrInvalidMetricsType = errors.New("invalid metrics type")
	ErrInvalidValueType   = errors.New("invalid value type")
	ErrMetricsNotFound    = errors.New("unknown this metric")
)
