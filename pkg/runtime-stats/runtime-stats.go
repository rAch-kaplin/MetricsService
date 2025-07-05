package runtimestats

import (
	"runtime"

	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
)

type MemRuntimeStat struct {
	Name string
	Type string
	Get  func(m *runtime.MemStats) float64
}

var MemRuntimeStats []MemRuntimeStat = []MemRuntimeStat{
	{
		Name: "Alloc",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.Alloc) },
	},
	{
		Name: "BuckHashSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.BuckHashSys) },
	},
	{
		Name: "Frees",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.Frees) },
	},
	{
		Name: "GCCPUFraction",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.GCCPUFraction) },
	},
	{
		Name: "HeapAlloc",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.HeapAlloc) },
	},
	{
		Name: "HeapIdle",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.HeapIdle) },
	},
	{
		Name: "HeapInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.HeapInuse) },
	},
	{
		Name: "HeapObjects",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.HeapObjects) },
	},
	{
		Name: "HeapReleased",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.HeapReleased) },
	},
	{
		Name: "HeapSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.HeapSys) },
	},
	{
		Name: "LastGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.LastGC) },
	},
	{
		Name: "Lookups",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.Lookups) },
	},
	{
		Name: "MCacheInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.MCacheInuse) },
	},
	{
		Name: "MCacheSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.MCacheSys) },
	},
	{
		Name: "MSpanInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.MSpanInuse) },
	},
	{
		Name: "MSpanSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.MSpanSys) },
	},
	{
		Name: "Mallocs",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.Mallocs) },
	},
	{
		Name: "NextGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.NextGC) },
	},
	{
		Name: "NumForcedGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.NumForcedGC) },
	},
	{
		Name: "NumGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.NumGC) },
	},
	{
		Name: "OtherSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.OtherSys) },
	},
	{
		Name: "PauseTotalNs",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.PauseTotalNs) },
	},
	{
		Name: "StackInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.StackInuse) },
	},
	{
		Name: "StackSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.StackSys) },
	},
	{
		Name: "Sys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.Sys) },
	},
	{
		Name: "TotalAlloc",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) float64 { return float64(m.TotalAlloc) },
	},
}
