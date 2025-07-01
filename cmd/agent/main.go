package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/MemStorage"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/logger"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second

	serverAddress = "http://localhost:8080"
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
		default:
			log.Error("ERROR: Unknown type for metric %s: %T", stat.Name, value)
			continue
		}

		if err := storage.UpdateMetric(metric); err != nil {
			log.Error("ERROR!")
		}
	}

	storage.UpdateMetric(mtr.NewCounter("PollCount", 1))
	storage.UpdateMetric(mtr.NewGauge("RandomValue", rand.Float64()))
}

func sendAllMetrics(client *http.Client, storage *ms.MemStorage) {
	gauges, counters := storage.GetAllMetrics()

	for name, value := range gauges {
		sendMetric(client, mtr.GaugeType, name, value)
	}

	for name, value := range counters {
		sendMetric(client, mtr.CounterType, name, value)
	}
}

func sendMetric(client *http.Client, mType string, mName string, mValue interface{}) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", serverAddress, mType, mName, mValue)

	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		log.Error("Error creating a request for %s: %v\n", mName, err)
		return
	}
	req.Header.Set("Content-Type", "text-plain")

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending metric: %s: %v\n", mName, err)
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		log.Error("Failed to read response body from %s: %v", url, err)
	}
}

func collectionLoop(storage *ms.MemStorage, interval time.Duration) {
	log.Debug("collectionLoop ...")
	for {
		UpdateAllMetrics(storage)
		time.Sleep(interval)
	}
}

func reportLoop(client *http.Client, storage *ms.MemStorage, interval time.Duration) {
	log.Debug("reportLoop ...")
	for {
		time.Sleep(interval)
		sendAllMetrics(client, storage)
	}
}

func main() {
	log.Init(log.DebugLevel, "logFile.log")
	defer log.Destroy()

	log.Debug("START AGENT>")
	storage := ms.NewMemStorage()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	go collectionLoop(storage, pollInterval)
	go reportLoop(client, storage, reportInterval)

	select {}
	log.Debug("END AGENT<")
}
