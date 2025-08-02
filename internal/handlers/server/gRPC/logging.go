package grpc

import (
	"context"
	"time"

	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// WithLogging is a function that logs the incoming request and the response.
//
// It logs the method, duration, and status of the request.
func WithLogging(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	response, err := handler(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("failed to handle request")
		return nil, err
	}

	duration := time.Since(start)
	st := status.Convert(err)

	log.Info().
		Str("grpc", "true").
		Str("method", info.FullMethod).
		Dur("duration", duration).
		Int("status", int(st.Code())).
		Msg("new request")

	return response, err
}
