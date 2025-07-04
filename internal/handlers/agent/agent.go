package agent

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"

	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/memstorage"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

type MemRuntimeStat struct {
	Name string
	Type string
	Get  func(m *runtime.MemStats) interface{}
}

var MemRuntimeStats []MemRuntimeStat = []MemRuntimeStat{
	{
		Name: "Alloc",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.Alloc },
	},
	{
		Name: "BuckHashSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.BuckHashSys },
	},
	{
		Name: "Frees",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.Frees },
	},
	{
		Name: "GCCPUFraction",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.GCCPUFraction },
	},
	{
		Name: "HeapAlloc",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.HeapAlloc },
	},
	{
		Name: "HeapIdle",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.HeapIdle },
	},
	{
		Name: "HeapInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.HeapInuse },
	},
	{
		Name: "HeapObjects",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.HeapObjects },
	},
	{
		Name: "HeapReleased",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.HeapReleased },
	},
	{
		Name: "HeapSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.HeapSys },
	},
	{
		Name: "LastGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.LastGC },
	},
	{
		Name: "Lookups",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.Lookups },
	},
	{
		Name: "MCacheInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.MCacheInuse },
	},
	{
		Name: "MCacheSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.MCacheSys },
	},
	{
		Name: "MSpanInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.MSpanInuse },
	},
	{
		Name: "MSpanSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.MSpanSys },
	},
	{
		Name: "Mallocs",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.Mallocs },
	},
	{
		Name: "NextGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.NextGC },
	},
	{
		Name: "NumForcedGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.NumForcedGC },
	},
	{
		Name: "NumGC",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.NumGC },
	},
	{
		Name: "OtherSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.OtherSys },
	},
	{
		Name: "PauseTotalNs",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.PauseTotalNs },
	},
	{
		Name: "StackInuse",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.StackInuse },
	},
	{
		Name: "StackSys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.StackSys },
	},
	{
		Name: "Sys",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.Sys },
	},
	{
		Name: "TotalAlloc",
		Type: mtr.GaugeType,
		Get:  func(m *runtime.MemStats) interface{} { return m.TotalAlloc },
	},
}

func UpdateAllMetrics(storage *ms.MemStorage) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for _, stat := range MemRuntimeStats {
		value := stat.Get(&memStats)
		var metric mtr.Metric

		switch v := value.(type) {
		case uint64:
			if stat.Type == mtr.GaugeType {
				metric = mtr.NewGauge(stat.Name, float64(v))
			}
		case float64:
			if stat.Type == mtr.GaugeType {
				metric = mtr.NewGauge(stat.Name, v)
			}
		case uint32:
			if stat.Type == mtr.GaugeType {
				metric = mtr.NewGauge(stat.Name, float64(v))
			}
		default:
			log.Error("ERROR: Unknown type for metric %s: %T", stat.Name, value)
			continue
		}

		if err := storage.UpdateMetric(metric); err != nil {
			log.Error("Failed to update metric %s: %v", stat.Name, err)
		}
	}

	if err := storage.UpdateMetric(mtr.NewCounter("PollCount", 1)); err != nil {
		log.Error("Failed to update PollCount metric: %v", err)
	}

	if err := storage.UpdateMetric(mtr.NewGauge("RandomValue", rand.Float64())); err != nil {
		log.Error("Failed to update RandomValue metric: %v", err)
	}
}

func sendAllMetrics(client *resty.Client, storage *ms.MemStorage) {
	gauges, counters := storage.GetAllMetrics()

	for name, value := range gauges {
		sendMetric(client, mtr.GaugeType, name, value)
	}

	for name, value := range counters {
		sendMetric(client, mtr.CounterType, name, value)
	}
}

func sendMetric(client *resty.Client, mType string, mName string, mValue interface{}) {
	res, err := client.R().
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"mType":  mType,
			"mName":  mName,
			"mValue": fmt.Sprintf("%v", mValue),
		}).
		Post("update/{mType}/{mName}/{mValue}")

	if err != nil {
		log.Error("Error creating a request for %s: %v", mName, err)
		return
	}

	if res.StatusCode() != http.StatusOK {
		log.Error("Server returned non-OK status for %s/%s: %d %s", mType, mName, res.StatusCode(), res.String())
	} else {
		log.Debug("Metric %s/%s sent successfully. Status: %d", mType, mName, res.StatusCode())
	}
}

func CollectionLoop(storage *ms.MemStorage, interval time.Duration) {
	log.Debug("collectionLoop ...")
	for {
		UpdateAllMetrics(storage)
		time.Sleep(interval)
	}
}

func ReportLoop(client *resty.Client, storage *ms.MemStorage, interval time.Duration) {
	log.Debug("reportLoop ...")
	for {
		time.Sleep(interval)
		sendAllMetrics(client, storage)
	}
}
