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
		val := stat.Get(&memStats)

		if err := storage.UpdateMetric(stat.Type, stat.Name, val); err != nil {
			log.Error().
				Err(err).
				Str("metric", stat.Name).
				Str("type", fmt.Sprintf("%T", val)).
				Msg("Failed to update runtime metric")
		}
	}

	if err := storage.UpdateMetric(mtr.CounterType, "PollCount", 1); err != nil {
		log.Error().Msgf("Failed to update PollCount metric: %v", err)
	}

	if err := storage.UpdateMetric(mtr.GaugeType, "RandomValue", rand.Float64()); err != nil {
		log.Error().Msgf("Failed to update RandomValue metric: %v", err)
	}
}

func sendAllMetrics(client *resty.Client, storage *ms.MemStorage) {
	allMetrics := storage.GetAllMetrics()

	for mType, innerMap := range allMetrics {
		for mName, metric := range innerMap {
			switch mType {
			case mtr.GaugeType:
				val, ok := metric.Value().(float64)
				if !ok {
					log.Error().Str("metric_name", mName).Str("metric_type", mType).
					Msg("Invalid metric value type")
					continue
				}
				sendMetric(client, mtr.GaugeType, mName, val)

			case mtr.CounterType:
				val, ok := metric.Value().(int64)
				if !ok {
					log.Error().Str("metric_name", mName).Str("metric_type", mType).
					Msg("Invalid metric value type")
					continue
				}
				sendMetric(client, mtr.CounterType, mName, val)
			}
		}
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
		log.Error().Msgf("Error creating a request for %s: %v", mName, err)
		return
	}

	if res.StatusCode() != http.StatusOK {
		log.Error().Msgf("Server returned non-OK status for %s/%s: %d %s", mType, mName, res.StatusCode(), res.String())
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
