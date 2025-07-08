package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mailru/easyjson"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/handlers/server"
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

		log.Debug().Msgf("update metric %s", stat.Name)
	}

	if err := storage.UpdateMetric(mtr.CounterType, "PollCount", int64(1)); err != nil {
		log.Error().Msgf("Failed to update PollCount metric: %v", err)
	}

	if err := storage.UpdateMetric(mtr.GaugeType, "RandomValue", rand.Float64()); err != nil {
		log.Error().Msgf("Failed to update RandomValue metric: %v", err)
	}
}

func sendAllMetrics(client *resty.Client, storage *ms.MemStorage) {
	allMetrics := storage.GetAllMetrics()

<<<<<<< HEAD
	for _, metric := range allMetrics {
		mType := metric.Type()
		mName := metric.Name()

		metricJSON := server.Metrics{
			ID:    mName,
			MType: mType,
		}

		switch mType {
		case mtr.GaugeType:
			val, ok := metric.Value().(float64)
			if !ok {
				log.Error().Str("metric_name", mName).Str("metric_type", mType).
					Msg("Invalid metric value type")
				continue
=======
	for mType, innerMap := range allMetrics {
		for mName, metric := range innerMap {

			metricJSON := server.Metrics{
				ID:    metric.Name(),
				MType: metric.Type(),
			}

			switch mType {
			case mtr.GaugeType:
				val, ok := metric.Value().(float64)
				if !ok {
					log.Error().Str("metric_name", mName).Str("metric_type", mType).
						Msg("Invalid metric value type")
					continue
				}

				metricJSON.Value = &val
				sendMetric(client, &metricJSON)

			case mtr.CounterType:
				val, ok := metric.Value().(int64)
				if !ok {
					log.Error().Str("metric_name", mName).Str("metric_type", mType).
						Msg("Invalid metric value type")
					continue
				}

				metricJSON.Delta = &val
				sendMetric(client, &metricJSON)
>>>>>>> c8d8784 (first easyjson-JSON version)
			}

			metricJSON.Value = &val
			sendMetric(client, &metricJSON)

		case mtr.CounterType:
			val, ok := metric.Value().(int64)
			if !ok {
				log.Error().Str("metric_name", mName).Str("metric_type", mType).
					Msg("Invalid metric value type")
				continue
			}

			metricJSON.Delta = &val
			sendMetric(client, &metricJSON)
		}
	}
}

func sendMetric(client *resty.Client, metricJSON *server.Metrics) {
	backoffSchedule := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	// buf, err := ConvertToGzipData(metricJSON)
	// if err != nil {
	// 	log.Error().Err(err).Msg("Failed to convert metric to gzip")
	// 	return
	// }

	for _, backoff := range backoffSchedule {
		res, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(metricJSON).
			Post("update/")

		if err != nil || res.StatusCode() != http.StatusOK {
		} else {
			break
		}

		time.Sleep(backoff)
	}
}

func ConvertToGzipData(metricJSON *server.Metrics) (*bytes.Buffer, error) {
	jsonData, err := easyjson.Marshal(*metricJSON)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal metricJSON")
		return nil, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer func() {
		err := gz.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close gzip writer")
		}
	}()

	_, err = gz.Write(jsonData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write gzip data")
		return nil, err
	}

	return &buf, nil
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
		sendAllMetrics(client, storage)
		time.Sleep(interval)
	}
}
