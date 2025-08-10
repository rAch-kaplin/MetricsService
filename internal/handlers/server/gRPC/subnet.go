package grpc

import (
	"context"
	"net"

	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// WithTrustedSubnet is a function that checks if the incoming request is from a trusted subnet.
//
// It parses the trusted subnet from the configuration, creates a subnet object,
// and returns a grpc.UnaryServerInterceptor that checks if the incoming request
// is from a trusted subnet. If the trusted subnet is not set, it returns a nil
// interceptor.
func WithTrustedSubnet(trustedSubnet string) grpc.UnaryServerInterceptor {
	var subnet *net.IPNet

	if trustedSubnet != "" {
		_, s, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse trusted subnet")
			return nil
		}

		subnet = s
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if subnet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Error().Msg("failed to get metadata")
			return nil, status.Errorf(codes.Internal, "failed to get metadata")
		}

		ipStr := md.Get("X-Real-IP")
		ip := net.ParseIP(ipStr[0])
		if ip == nil {
			log.Error().Msg("failed to parse ip")
			return nil, status.Errorf(codes.Internal, "failed to parse ip")
		}

		if !subnet.Contains(ip) {
			log.Error().Msg("ip is not in trusted subnet")
			return nil, status.Errorf(codes.PermissionDenied, "ip is not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
