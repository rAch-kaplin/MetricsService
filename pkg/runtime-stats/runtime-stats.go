package runtimestats

import (
	"runtime"

	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
)

type MemRuntimeStat struct {
	Name string
	Type string
	Get  func(m *runtime.MemStats) any
}

var MemRuntimeStats []MemRuntimeStat = []MemRuntimeStat{
	{
		Name: "Alloc",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.Alloc },
	},
	{
		Name: "BuckHashSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.BuckHashSys },
	},
	{
		Name: "Frees",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.Frees },
	},
	{
		Name: "GCCPUFraction",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.GCCPUFraction },
	},
	{
		Name: "HeapAlloc",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.HeapAlloc },
	},
	{
		Name: "HeapIdle",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.HeapIdle },
	},
	{
		Name: "HeapInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.HeapInuse },
	},
	{
		Name: "HeapObjects",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.HeapObjects },
	},
	{
		Name: "HeapReleased",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.HeapReleased },
	},
	{
		Name: "HeapSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.HeapSys },
	},
	{
		Name: "LastGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.LastGC },
	},
	{
		Name: "Lookups",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.Lookups },
	},
	{
		Name: "MCacheInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.MCacheInuse },
	},
	{
		Name: "MCacheSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.MCacheSys },
	},
	{
		Name: "MSpanInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.MSpanInuse },
	},
	{
		Name: "MSpanSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.MSpanSys },
	},
	{
		Name: "Mallocs",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.Mallocs },
	},
	{
		Name: "NextGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.NextGC },
	},
	{
		Name: "NumForcedGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.NumForcedGC },
	},
	{
		Name: "NumGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.NumGC },
	},
	{
		Name: "OtherSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.OtherSys },
	},
	{
		Name: "PauseTotalNs",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.PauseTotalNs },
	},
	{
		Name: "StackInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.StackInuse },
	},
	{
		Name: "StackSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.StackSys },
	},
	{
		Name: "Sys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.Sys },
	},
	{
		Name: "TotalAlloc",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return m.TotalAlloc },
	},
}
