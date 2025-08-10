// @Title Agent Handlers API
// @Description This package implements the core logic of the metrics collection agent.
// The collected metrics include:
//
// Gauge Metrics:
// - Alloc: Bytes of allocated heap objects.
// - BuckHashSys: Bytes used by the profiling bucket hash table.
// - Frees: Total number of frees.
// - GCCPUFraction: Fraction of CPU time used by GC.
// - GCSys: Bytes used for garbage collection system metadata.
// - HeapAlloc: Bytes allocated and still in use on the heap.
// - HeapIdle: Bytes in idle spans.
// - HeapInuse: Bytes in in-use spans.
// - HeapObjects: Number of allocated heap objects.
// - HeapReleased: Bytes of physical memory returned to the OS.
// - HeapSys: Bytes of heap memory obtained from the OS.
// - LastGC: Time the last garbage collection finished, as nanoseconds since epoch.
// - Lookups: Number of pointer lookups.
// - MCacheInuse: Bytes used for mcache structures.
// - MCacheSys: Bytes obtained from the OS for mcache structures.
// - MSpanInuse: Bytes used for mspan structures.
// - MSpanSys: Bytes obtained from the OS for mspan structures.
// - Mallocs: Total number of mallocs.
// - NextGC: Target heap size for the next GC cycle.
// - NumForcedGC: Number of forced GCs.
// - NumGC: Number of completed GC cycles.
// - OtherSys: Bytes used for other system allocations.
// - PauseTotalNs: Cumulative nanoseconds in GC stop-the-world pauses.
// - StackInuse: Bytes used by stack spans.
// - StackSys: Bytes obtained from the OS for stack spans.
// - Sys: Total bytes obtained from the OS.
// - TotalAlloc: Cumulative bytes allocated for heap objects.
// - RandomValue: A random value.
// - TotalMemory: Total memory available to the process.
// - FreeMemory: Free memory available to the process.
// - CPUutilization1: CPU utilization percentage.
//
// Counter Metrics:
// - PollCount: Number of times the agent has polled metrics.
//
// These metrics are collected periodically by the agent and sent to a remote server using a metrics API
//
// @Author rAch-kaplin
// @Version 1.0.0
// @Since 2025-07-29

package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mailru/easyjson"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/agent"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	pb "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/grpc-metrics"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/hash"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	rt "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/runtime-stats"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
	worker "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/worker-pool"
)

// Agent is a struct that contains the use case for the agent.
type Agent struct {
	Usecase *agent.AgentUsecase
}

// NewAgent is a function that creates a new agent.
func NewAgent(uc *agent.AgentUsecase) *Agent {
	return &Agent{Usecase: uc}
}

// @Title UpdateAllMetrics
// @Description Update all metrics
// @Tags metrics
// @Produces text/plain
// @Success 200 {string} string "Metrics updated successfully"
// @Failure 500 {string} string "Internal server error"
func (ag *Agent) UpdateAllMetrics(ctx context.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for _, stat := range rt.MemRuntimeStats {
		val := stat.Get(&memStats)

		if err := ag.Usecase.UpdateMetric(ctx, stat.Type, stat.Name, val); err != nil {
			log.Error().
				Err(err).
				Str("metric", stat.Name).
				Str("type", fmt.Sprintf("%T", val)).
				Msg("Failed to update runtime metric")
		}

		log.Debug().Msgf("update metric %s", stat.Name)
	}

	if err := ag.Usecase.UpdateMetric(ctx, models.CounterType, "PollCount", int64(1)); err != nil {
		log.Error().Msgf("Failed to update PollCount metric: %v", err)
	}

	if err := ag.Usecase.UpdateMetric(ctx, models.GaugeType, "RandomValue", rand.Float64()); err != nil {
		log.Error().Msgf("Failed to update RandomValue metric: %v", err)
	}

	v, _ := mem.VirtualMemory()
	if err := ag.Usecase.UpdateMetric(ctx, models.GaugeType, "TotalMemory", float64(v.Total)); err != nil {
		log.Error().Msgf("Failed to update TotalMemory metric: %v", err)
	}

	if err := ag.Usecase.UpdateMetric(ctx, models.GaugeType, "FreeMemory", float64(v.Free)); err != nil {
		log.Error().Msgf("Failed to update FreeMemory metric: %v", err)
	}

	percent, _ := cpu.Percent(0, false)
	if err := ag.Usecase.UpdateMetric(ctx, models.GaugeType, "CPUutilization1", percent[0]); err != nil {
		log.Error().Msgf("Failed to update CPUutilization1 metric: %v", err)
	}
}

// @Title SendAllMetrics
// @Description Send all metrics to the server
// @Tags metrics
// @Produces text/plain
// @Success 200 {string} string "Metrics sent successfully"
// @Failure 500 {string} string "Internal server error"
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

// @Title sendBatch
// @Description Send a batch of metrics to the server
// @Tags metrics
// @Produces text/plain
// @Success 200 {string} string "Metrics sent successfully"
// @Failure 500 {string} string "Internal server error"
func sendBatch(client *resty.Client, metrics []serialize.Metric, key string) {
	// Create a backoff schedule for the agent.
	backoffSchedule := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	// Convert the metrics to gzip data.
	buf, ok, err := ConvertToGzipData(metrics)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert metric to gzip")
		return
	}

	// Create a hash for the metrics.
	var h string
	if key != "" {
		hashBytes, err := hash.GetHash([]byte(key), buf.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("can't get hash")
			return
		}

		h = hex.EncodeToString(hashBytes)
	}

	// Get the outbound IP of the machine.
	ip, err := GetOutboundIP()
	if err != nil {
		log.Error().Err(err).Msg("can't get outbound ip")
		return
	}

	// Send the metrics to the server with a backoff schedule.
	for _, backoff := range backoffSchedule {
		req := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("X-Real-IP", ip.String()).
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

// This func is used to get the outbound ip of the machine
func GetOutboundIP() (net.IP, error) {
	// Create a UDP connection to a known IP address
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = conn.Close()
	}()

	return conn.LocalAddr().(*net.UDPAddr).IP, nil
}

// @Title ConvertToGzipData
// @Description Convert metrics to gzip data
// @Tags metrics
// @Produces text/plain
// @Success 200 {string} string "Metrics converted successfully"
// @Failure 500 {string} string "Internal server error"
func ConvertToGzipData(metrics serialize.MetricsList) (*bytes.Buffer, bool, error) {
	var jsonBuf bytes.Buffer

	// Marshal the metrics to JSON.
	_, err := easyjson.MarshalToWriter(metrics, &jsonBuf)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal metrics")
		return nil, false, err
	}

	// If the metrics is less than 1024 bytes, we don't need to compress it.
	if jsonBuf.Len() <= 1024 {
		return &jsonBuf, false, nil
	}

	var buf bytes.Buffer

	// Create a gzip writer.
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

	// Write the metrics to the gzip writer.
	_, err = gz.Write(jsonBuf.Bytes())
	if err != nil {
		log.Error().Err(err).Msg("Failed to write gzip data")
		return nil, false, err
	}

	return &buf, true, nil
}

// This func is used to collect metrics from the system every pollInterval seconds.
func CollectMetrics(ctx context.Context, ag *Agent, pollInterval int) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ag.UpdateAllMetrics(ctx)
		}
	}
}

// This func is used to send metrics to the server every reportInterval seconds.
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
			wp.AddTask(func(ctx context.Context) error {
				ag.SendAllMetrics(ctx, client, key)
				return nil
			})
		}
	}

}

// This func is used to send metrics to the server using gRPC.
func (ag *Agent) SendAllMetricsGRPC(ctx context.Context, client pb.MetricsServiceClient, key string) {
	// Get all metrics from the use case.
	allMetrics, err := ag.Usecase.GetAllMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to Get metrics")
	}

	// Convert the metrics to proto.
	metricsToProto, err := converter.ConvertToProtoMetrics(allMetrics)
	if err != nil {
		_ = status.Errorf(codes.Internal, "failed to convert metrics to proto")
		log.Error().Err(err).Msg("failed to convert metrics to proto")
		return
	}

	// Marshal the metrics to proto.
	data, err := proto.Marshal(&pb.UpdateMetricsRequest{
		Metrics: metricsToProto,
	})
	if err != nil {
		_ =status.Errorf(codes.Internal, "failed to marshal metrics to proto")
		log.Error().Err(err).Msg("failed to marshal metrics to proto")
		return
	}

	// Create a hash for the metrics.
	hashBytes, err := hash.GetHash([]byte(key), data)
	if err != nil {
		_ = status.Errorf(codes.Internal, "failed to get hash")
		log.Error().Err(err).Msg("failed to get hash")
		return
	}

	h := hex.EncodeToString(hashBytes)

	// Create a metadata for the metrics.
	md := metadata.New(map[string]string{"HashSHA256": h})

	// Create a context for the metrics.
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Send the metrics to the server.
	_, err = client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{
		Metrics: metricsToProto,
	})
	if err != nil {
		_ = status.Errorf(codes.Internal, "failed to send metrics")
		log.Error().Err(err).Msg("failed to send metrics")
		return
	}
}

// This func is used to send metrics to the server using gRPC every reportInterval seconds.
func SendMetricsGRPC(ctx context.Context,
	ag *Agent,
	client pb.MetricsServiceClient,
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
			wp.AddTask(func(ctx context.Context) error {
				ag.SendAllMetricsGRPC(ctx, client, key)
				return nil
			})
		}
	}
}
