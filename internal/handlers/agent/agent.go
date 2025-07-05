package agent

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"

	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	rt "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/runtime-stats"
)

func UpdateAllMetrics(storage *ms.MemStorage) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for _, stat := range rt.MemRuntimeStats {
		if err := storage.UpdateMetric(mtr.NewGauge(stat.Name, stat.Get(&memStats))); err != nil {
			log.Error().
				Err(err).
				Str("metric", stat.Name).
				Msg("Failed to update metric")
		}
	}

	if err := storage.UpdateMetric(mtr.NewCounter("PollCount", 1)); err != nil {
		log.Error().
			Err(err).
			Str("metric", "PollCount").
			Msg("Failed to update PollCount")
	}

	if err := storage.UpdateMetric(mtr.NewGauge("RandomValue", rand.Float64())); err != nil {
		log.Error().
			Err(err).
			Str("metric", "RandomValue").
			Msg("Failed to update RandomValue")
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
		log.Error().
			Err(err).
			Str("metric", mName).
			Msg("Error sending request")
		return
	}

	if res.StatusCode() != http.StatusOK {
		log.Error().
			Str("type", mType).
			Str("name", mName).
			Int("status", res.StatusCode()).
			Str("response", res.String()).
			Msg("Non-OK response from server")
	} else {
		log.Debug().
			Str("type", mType).
			Str("name", mName).
			Int("status", res.StatusCode()).
			Msg("Metric sent successfully")
	}
}

func CollectionLoop(storage *ms.MemStorage, interval time.Duration) {
	log.Debug().Msg("collectionLoop ...")
	for {
		UpdateAllMetrics(storage)
		time.Sleep(interval)
	}
}

func ReportLoop(client *resty.Client, storage *ms.MemStorage, interval time.Duration) {
	log.Debug().Msg("reportLoop ...")
	for {
		time.Sleep(interval)
		sendAllMetrics(client, storage)
	}
}
