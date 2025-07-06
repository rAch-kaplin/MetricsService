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

func toFloat64(val any) (float64, bool) {
    switch v := val.(type) {
    case float64:
        return v, true
    case uint64:
        return float64(v), true
    case uint32:
        return float64(v), true
    case int:
        return float64(v), true
    case int64:
        return float64(v), true
    default:
        return 0, false
    }
}

func UpdateAllMetrics(storage *ms.MemStorage) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for _, stat := range rt.MemRuntimeStats {
		val := stat.Get(&memStats)

		switch stat.Type {
		case mtr.GaugeType:
			value, ok := toFloat64(val)
			if !ok {
				log.Error().
					Str("metric", stat.Name).
					Str("type", fmt.Sprintf("%T", val)).
					Msg("Failed to convert metric value to float64")
				continue
			}

			if err := storage.UpdateMetric(mtr.NewGauge(stat.Name, value)); err != nil {
            log.Error().
                Err(err).
                Str("metric", stat.Name).
                Msg("Failed to update metric")
        	}
		default:
			 log.Error().
                Str("metric", stat.Name).
                Str("type", fmt.Sprintf("%T", val)).
                Msg("Unsupported metric type")
		}
	}

	if err := storage.UpdateMetric(mtr.NewCounter("PollCount", 1)); err != nil {
		log.Error().Msgf("Failed to update PollCount metric: %v", err)
	}

	if err := storage.UpdateMetric(mtr.NewGauge("RandomValue", rand.Float64())); err != nil {
		log.Error().Msgf("Failed to update RandomValue metric: %v", err)
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
		log.Error().Msgf("Error creating a request for %s: %v", mName, err)
		return
	}

	if res.StatusCode() != http.StatusOK {
		log.Error().Msgf("Server returned non-OK status for %s/%s: %d %s", mType, mName, res.StatusCode(), res.String())
	} else {
		log.Debug().Msgf("Metric %s/%s sent successfully. Status: %d", mType, mName, res.StatusCode())
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
