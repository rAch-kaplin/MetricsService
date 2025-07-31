package grpc

import (
	"context"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/ping"
	srvUsecase "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/server"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	pb "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/grpc-metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedMetricsServiceServer
	MetricUsecase *srvUsecase.MetricUsecase
	PingUsecase   *ping.PingUsecase
}

func NewServer(uc *srvUsecase.MetricUsecase, puc *ping.PingUsecase) *Server {
	return &Server{
		MetricUsecase: uc,
		PingUsecase:   puc,
	}
}

func (s *Server) GetMetric(ctx context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	metric, err := s.MetricUsecase.GetMetric(ctx, req.Type, req.Id)
	if err != nil {
		log.Error().Err(err).Msg("failed to get metric")
		return nil, status.Errorf(codes.NotFound, "failed to get metric: %v", err)
	}

	protoMetric := converter.ConvertToProtoMetrics([]models.Metric{metric})

	return &pb.GetMetricResponse{
		Metric: protoMetric[0],
	}, nil
}

func (s *Server) GetAllMetrics(ctx context.Context) (*pb.GetAllMetricsResponse, error) {
	metrics, err := s.MetricUsecase.GetAllMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get all metrics")
		return nil, status.Errorf(codes.NotFound, "failed to get all metrics: %v", err)
	}

	protoMetrics := converter.ConvertToProtoMetrics(metrics)

	return &pb.GetAllMetricsResponse{
		Metrics: protoMetrics,
	}, nil
}
