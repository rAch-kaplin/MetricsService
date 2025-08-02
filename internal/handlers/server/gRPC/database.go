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

// WithDataBase is a function that checks if the database is reachable.
//
// It creates a context with a timeout of 1 second and pings the database.
// If the database is not reachable, it returns an Internal error.
// If the database is reachable, it returns handler.
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
