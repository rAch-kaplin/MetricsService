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
	"google.golang.org/protobuf/types/known/emptypb"
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

	protoMetric, err := converter.ConvertToProtoMetrics([]models.Metric{metric})
	if err != nil {
		log.Error().Err(err).Msg("failed to convert metric to proto")
		return nil, status.Errorf(codes.Internal, "failed to convert metric to proto: %v", err)
	}

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

	protoMetrics, err := converter.ConvertToProtoMetrics(metrics)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert metrics to proto")
		return nil, status.Errorf(codes.Internal, "failed to convert metrics to proto: %v", err)
	}

	return &pb.GetAllMetricsResponse{
		Metrics: protoMetrics,
	}, nil
}

func (s *Server) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*emptypb.Empty, error) {
	metric := req.Metric

	if metric.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "metric id is required")
	}

	if err := s.MetricUsecase.UpdateMetric(ctx, metric.MType, metric.Id, metric.MetricValue); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update metric: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*emptypb.Empty, error) {
	protoMetrics := req.Metrics

	if len(protoMetrics) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "metrics are required")
	}

	metrics, err := converter.ConvertFromProtoToMetrics(protoMetrics)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert metrics")
		return nil, status.Errorf(codes.Internal, "failed to convert metrics: %v", err)
	}

	if err := s.MetricUsecase.UpdateMetricList(ctx, metrics); err != nil {
		log.Error().Err(err).Msg("failed to update metrics")
		return nil, status.Errorf(codes.Internal, "failed to update metrics: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) PingHandler(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.PingUsecase.Check(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ping")
		return nil, status.Errorf(codes.Internal, "failed to ping: %v", err)
	}

	return &emptypb.Empty{}, nil
}
