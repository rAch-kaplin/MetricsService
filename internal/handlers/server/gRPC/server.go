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

// Server provides the gRPC implementation of the MetricsServiceServer.
//
// This struct acts as a bridge between the gRPC layer and the application
// business logic (use cases). It embeds pb.UnimplementedMetricsServiceServer
// so that you only need to implement the RPC methods that are actually used.
type Server struct {
	pb.UnimplementedMetricsServiceServer
	MetricUsecase *srvUsecase.MetricUsecase
	PingUsecase   *ping.PingUsecase
}

// NewServer creates a new Server with the given use cases.
func NewServer(uc *srvUsecase.MetricUsecase, puc *ping.PingUsecase) *Server {
	return &Server{
		MetricUsecase: uc,
		PingUsecase:   puc,
	}
}

// GetMetric implements the GetMetric RPC method.
//
// It retrieves a single metric by its type and name, converts it to a protobuf
// message, and returns it. If the metric is not found, it returns a NotFound
// error. If there is an internal error, it returns an Internal error.
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

// GetAllMetrics implements the GetAllMetrics RPC method.
//
// It retrieves all metrics from the use case, converts them to protobuf
// messages, and returns them. If there is an internal error, it returns an
// Internal error.
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

// UpdateMetric implements the UpdateMetric RPC method.
//
// It updates a single metric by its type and name.
// If there is an internal error, it returns an Internal error.
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

// UpdateMetrics implements the UpdateMetrics RPC method.
//
// It updates a list of metrics.
// If there is an internal error, it returns an Internal error.
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

// PingHandler implements the PingHandler RPC method.
//
// It checks if the database is reachable.
// If there is an internal error, it returns an Internal error.
func (s *Server) PingHandler(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.PingUsecase.Check(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ping")
		return nil, status.Errorf(codes.Internal, "failed to ping: %v", err)
	}

	return &emptypb.Empty{}, nil
}
