package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mailru/easyjson"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	rt "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/runtime-stats"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
)

func UpdateAllMetrics(ctx context.Context, storage col.Collector) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for _, stat := range rt.MemRuntimeStats {
		val := stat.Get(&memStats)

		if err := storage.UpdateMetric(ctx, stat.Type, stat.Name, val); err != nil {
			log.Error().
				Err(err).
				Str("metric", stat.Name).
				Str("type", fmt.Sprintf("%T", val)).
				Msg("Failed to update runtime metric")
		}

		log.Debug().Msgf("update metric %s", stat.Name)
	}

	if err := storage.UpdateMetric(ctx, models.CounterType, "PollCount", int64(1)); err != nil {
		log.Error().Msgf("Failed to update PollCount metric: %v", err)
	}

	if err := storage.UpdateMetric(ctx, models.GaugeType, "RandomValue", rand.Float64()); err != nil {
		log.Error().Msgf("Failed to update RandomValue metric: %v", err)
	}
}

func (uc *AgentUsecase) SendAllMetrics(ctx context.Context, client *resty.Client) {
	allMetrics, err := uc.GetAllMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to Get metrics")
	}

	log.Debug().Msgf("send all metrics: got %d metrics", len(allMetrics))
	for _, metric := range allMetrics {
		jsonMetric, err := uc.GetMetricJSON(ctx, metric)
		if err != nil {
			log.Error().Err(err).Msg("failed to get json metric")
		}

		sendMetric(client, jsonMetric)
	}
}

func sendMetric(client *resty.Client, metricJSON *serialize.Metric) {
	backoffSchedule := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	log.Debug().Msg("sendMetric")
	buf, ok, err := ConvertToGzipData(metricJSON)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert metric to gzip")
		return
	}

	for _, backoff := range backoffSchedule {
		req := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(buf)

		if ok {
			req.SetHeader("Content-Encoding", "gzip")
		}

		res, err := req.Post("update/")

		if err != nil || res.StatusCode() != http.StatusOK {
		} else {
			break
		}

		time.Sleep(backoff)
	}
}

func ConvertToGzipData(metricJSON *serialize.Metric) (*bytes.Buffer, bool, error) {
	jsonData, err := easyjson.Marshal(*metricJSON)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal metricJSON")
		return nil, false, err
	}

	var buf bytes.Buffer
	if len(jsonData) <= 1024 {
		buf.Write(jsonData)
		return &buf, false, nil
	}

	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create gzip writer")
		return nil, false, err
	}
	defer func() {
		err := gz.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close gzip writer")
		}
	}()

	_, err = gz.Write(jsonData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write gzip data")
		return nil, false, err
	}

	return &buf, true, nil
}
