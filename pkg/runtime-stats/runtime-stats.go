package runtimestats

import (
	"runtime"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
)

type MemRuntimeStat struct {
	Name string
	Type string
	Get  func(m *runtime.MemStats) any
}

var MemRuntimeStats []MemRuntimeStat = []MemRuntimeStat{
	{
		Name: "Alloc",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.Alloc) },
	},
	{
		Name: "BuckHashSys",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.BuckHashSys) },
	},
	{
		Name: "Frees",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.Frees) },
	},
	{
		Name: "GCCPUFraction",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.GCCPUFraction) },
	},
	{
		Name: "GCSys",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.GCSys) },
	},
	{
		Name: "HeapAlloc",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapAlloc) },
	},
	{
		Name: "HeapIdle",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapIdle) },
	},
	{
		Name: "HeapInuse",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapInuse) },
	},
	{
		Name: "HeapObjects",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapObjects) },
	},
	{
		Name: "HeapReleased",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapReleased) },
	},
	{
		Name: "HeapSys",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapSys) },
	},
	{
		Name: "LastGC",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.LastGC) },
	},
	{
		Name: "Lookups",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.Lookups) },
	},
	{
		Name: "MCacheInuse",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.MCacheInuse) },
	},
	{
		Name: "MCacheSys",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.MCacheSys) },
	},
	{
		Name: "MSpanInuse",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.MSpanInuse) },
	},
	{
		Name: "MSpanSys",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.MSpanSys) },
	},
	{
		Name: "Mallocs",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.Mallocs) },
	},
	{
		Name: "NextGC",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.NextGC) },
	},
	{
		Name: "NumForcedGC",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.NumForcedGC) },
	},
	{
		Name: "NumGC",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.NumGC) },
	},
	{
		Name: "OtherSys",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.OtherSys) },
	},
	{
		Name: "PauseTotalNs",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.PauseTotalNs) },
	},
	{
		Name: "StackInuse",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.StackInuse) },
	},
	{
		Name: "StackSys",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.StackSys) },
	},
	{
		Name: "Sys",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.Sys) },
	},
	{
		Name: "TotalAlloc",
		Type: models.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.TotalAlloc) },
	},
}
