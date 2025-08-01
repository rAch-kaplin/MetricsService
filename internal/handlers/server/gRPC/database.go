package grpc

import (
	"context"
	"database/sql"
	"time"

	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WithDataBase(db *sql.DB) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Error().Err(err).Msg("database ping failed")
			return nil, status.Errorf(codes.Internal, "database ping failed")
		}
		log.Info().Msg("database ping successful")

		return handler(ctx, req)
	}
}
