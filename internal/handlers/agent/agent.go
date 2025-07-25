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

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/agent"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	rt "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/runtime-stats"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
)

type Agent struct {
	Usecase *agent.AgentUsecase
}

func NewAgent(uc *agent.AgentUsecase) *Agent {
	return &Agent{Usecase: uc}
}

func UpdateAllMetrics(ctx context.Context, storage agent.MetricUpdater) {
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

func (ag *Agent) SendAllMetrics(ctx context.Context, client *resty.Client) {
	allMetrics, err := ag.Usecase.GetAllMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to Get metrics")
	}

	metricsToSend, err := converter.ConvertToSerialization(allMetrics)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert metrics to serialization")
	}

	if len(metricsToSend) > 0 {
		sendBatch(client, metricsToSend)
	}

	log.Info().Int("count", len(metricsToSend)).Msg("Sending metrics batch")
}

func sendBatch(client *resty.Client, metrics []serialize.Metric) {
	backoffSchedule := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	log.Debug().Msg("sendMetric")
	buf, ok, err := ConvertToGzipData(metrics)
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

		res, err := req.Post("updates/")

		if err != nil || res.StatusCode() != http.StatusOK {
		} else {
			break
		}

		time.Sleep(backoff)
	}
}

func ConvertToGzipData(metrics serialize.MetricsList) (*bytes.Buffer, bool, error) {
	var jsonBuf bytes.Buffer

	_, err := easyjson.MarshalToWriter(metrics, &jsonBuf)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal metrics")
		return nil, false, err
	}

	if jsonBuf.Len() <= 1024 {
		return &jsonBuf, false, nil
	}

	var buf bytes.Buffer
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

	_, err = gz.Write(jsonBuf.Bytes())
	if err != nil {
		log.Error().Err(err).Msg("Failed to write gzip data")
		return nil, false, err
	}

	return &buf, true, nil
}
