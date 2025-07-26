package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mailru/easyjson"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	repo "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/repository"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/agent"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/hash"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	rt "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/runtime-stats"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
	worker "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/worker-pool"
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

	v, _ := mem.VirtualMemory()
	if err := storage.UpdateMetric(ctx, models.GaugeType, "TotalMemory", float64(v.Total)); err != nil {
		log.Error().Msgf("Failed to update TotalMemory metric: %v", err)
	}

	if err := storage.UpdateMetric(ctx, models.GaugeType, "FreeMemory", float64(v.Free)); err != nil {
		log.Error().Msgf("Failed to update FreeMemory metric: %v", err)
	}

	percent, _ := cpu.Percent(0, false)
	if err := storage.UpdateMetric(ctx, models.GaugeType, "CPUutilization1", percent[0]); err != nil {
		log.Error().Msgf("Failed to update CPUutilization1 metric: %v", err)
	}
}

func (ag *Agent) SendAllMetrics(ctx context.Context, client *resty.Client, key string) {
	allMetrics, err := ag.Usecase.GetAllMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to Get metrics")
	}

	metricsToSend, err := converter.ConvertToSerialization(allMetrics)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert metrics to serialization")
	}

	if len(metricsToSend) > 0 {
		sendBatch(client, metricsToSend, key)
	}

	log.Info().Int("count", len(metricsToSend)).Msg("Sending metrics batch")
}

func sendBatch(client *resty.Client, metrics []serialize.Metric, key string) {
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

	var h string
	if key != "" {
		hashBytes, err := hash.GetHash([]byte(key), buf.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("can't get hash")
			return
		}

		h = hex.EncodeToString(hashBytes)
	}

	for _, backoff := range backoffSchedule {
		req := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(buf.Bytes())

		if ok {
			req.SetHeader("Content-Encoding", "gzip")
		}

		if h != "" {
			req.SetHeader("HashSHA256", h)
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

func CollectMetrics(ctx context.Context, storage *repo.MemStorage, pollInterval int) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			UpdateAllMetrics(ctx, storage)
		}
	}
}

func SendMetrics(ctx context.Context,
	ag *Agent,
	client *resty.Client,
	wp *worker.WorkerPool,
	reportInterval int,
	key string) {

	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			wp.AddTask(&worker.Task{
				Execute: func() {
					ag.SendAllMetrics(ctx, client, key)
				},
			})
		}
	}

}
