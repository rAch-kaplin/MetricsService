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
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.Alloc) },
=======
		Get:  func(m *runtime.MemStats) any { return m.Alloc },
>>>>>>> main
	},
	{
		Name: "BuckHashSys",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.BuckHashSys) },
=======
		Get:  func(m *runtime.MemStats) any { return m.BuckHashSys },
>>>>>>> main
	},
	{
		Name: "Frees",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.Frees) },
=======
		Get:  func(m *runtime.MemStats) any { return m.Frees },
>>>>>>> main
	},
	{
		Name: "GCCPUFraction",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.GCCPUFraction) },
	},
	{
		Name: "GCSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) any { return float64(m.GCSys) },
=======
		Get:  func(m *runtime.MemStats) any { return m.GCCPUFraction },
>>>>>>> main
	},
	{
		Name: "HeapAlloc",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapAlloc) },
=======
		Get:  func(m *runtime.MemStats) any { return m.HeapAlloc },
>>>>>>> main
	},
	{
		Name: "HeapIdle",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapIdle) },
=======
		Get:  func(m *runtime.MemStats) any { return m.HeapIdle },
>>>>>>> main
	},
	{
		Name: "HeapInuse",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapInuse) },
=======
		Get:  func(m *runtime.MemStats) any { return m.HeapInuse },
>>>>>>> main
	},
	{
		Name: "HeapObjects",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapObjects) },
=======
		Get:  func(m *runtime.MemStats) any { return m.HeapObjects },
>>>>>>> main
	},
	{
		Name: "HeapReleased",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapReleased) },
=======
		Get:  func(m *runtime.MemStats) any { return m.HeapReleased },
>>>>>>> main
	},
	{
		Name: "HeapSys",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.HeapSys) },
=======
		Get:  func(m *runtime.MemStats) any { return m.HeapSys },
>>>>>>> main
	},
	{
		Name: "LastGC",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.LastGC) },
=======
		Get:  func(m *runtime.MemStats) any { return m.LastGC },
>>>>>>> main
	},
	{
		Name: "Lookups",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.Lookups) },
=======
		Get:  func(m *runtime.MemStats) any { return m.Lookups },
>>>>>>> main
	},
	{
		Name: "MCacheInuse",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.MCacheInuse) },
=======
		Get:  func(m *runtime.MemStats) any { return m.MCacheInuse },
>>>>>>> main
	},
	{
		Name: "MCacheSys",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.MCacheSys) },
=======
		Get:  func(m *runtime.MemStats) any { return m.MCacheSys },
>>>>>>> main
	},
	{
		Name: "MSpanInuse",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.MSpanInuse) },
=======
		Get:  func(m *runtime.MemStats) any { return m.MSpanInuse },
>>>>>>> main
	},
	{
		Name: "MSpanSys",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.MSpanSys) },
=======
		Get:  func(m *runtime.MemStats) any { return m.MSpanSys },
>>>>>>> main
	},
	{
		Name: "Mallocs",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.Mallocs) },
=======
		Get:  func(m *runtime.MemStats) any { return m.Mallocs },
>>>>>>> main
	},
	{
		Name: "NextGC",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.NextGC) },
=======
		Get:  func(m *runtime.MemStats) any { return m.NextGC },
>>>>>>> main
	},
	{
		Name: "NumForcedGC",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.NumForcedGC) },
=======
		Get:  func(m *runtime.MemStats) any { return m.NumForcedGC },
>>>>>>> main
	},
	{
		Name: "NumGC",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.NumGC) },
=======
		Get:  func(m *runtime.MemStats) any { return m.NumGC },
>>>>>>> main
	},
	{
		Name: "OtherSys",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.OtherSys) },
=======
		Get:  func(m *runtime.MemStats) any { return m.OtherSys },
>>>>>>> main
	},
	{
		Name: "PauseTotalNs",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.PauseTotalNs) },
=======
		Get:  func(m *runtime.MemStats) any { return m.PauseTotalNs },
>>>>>>> main
	},
	{
		Name: "StackInuse",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.StackInuse) },
=======
		Get:  func(m *runtime.MemStats) any { return m.StackInuse },
>>>>>>> main
	},
	{
		Name: "StackSys",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.StackSys) },
=======
		Get:  func(m *runtime.MemStats) any { return m.StackSys },
>>>>>>> main
	},
	{
		Name: "Sys",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.Sys) },
=======
		Get:  func(m *runtime.MemStats) any { return m.Sys },
>>>>>>> main
	},
	{
		Name: "TotalAlloc",
		Type: mtr.GaugeType,
<<<<<<< HEAD
		Get:  func(m *runtime.MemStats) any { return float64(m.TotalAlloc) },
=======
		Get:  func(m *runtime.MemStats) any { return m.TotalAlloc },
>>>>>>> main
	},
}
