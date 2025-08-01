package grpc

import (
	"context"
	"encoding/hex"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/hash"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func WithHashing(key []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "failed to get metadata")
		}

		hashes := md.Get("HashSHA256")
		if len(hashes) == 0 {
			return handler(ctx, req)
		}

		decoded, err := hex.DecodeString(hashes[0])
		if err != nil {
			log.Error().Err(err).Msg("failed to decode hash")
			return nil, status.Errorf(codes.Internal, "failed to decode hash: %v", err)
		}

		if msg, ok := req.(proto.Message); ok {
			body, err := proto.Marshal(msg)
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal request")
				return nil, status.Errorf(codes.Internal, "failed to marshal request: %v", err)
			}

			valid := hash.CheckHash(key, body, decoded)
			if !valid {
				log.Error().Msg("invalid hash message")
				return nil, status.Errorf(codes.Internal, "invalid hash message")
			}
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		if msg, ok := resp.(proto.Message); ok {
			body, err := proto.Marshal(msg)
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal response")
				return nil, status.Errorf(codes.Internal, "failed to marshal response: %v", err)
			}
			if len(body) > 0 {
				newHash, err := hash.GetHash(key, body)
				if err != nil {
					log.Error().Err(err).Msg("failed to get new hash")
					return nil, status.Errorf(codes.Internal, "failed to get new hash: %v", err)
				}

				md := metadata.Pairs("HashSHA256", hex.EncodeToString(newHash))
				if err := grpc.SetHeader(ctx, md); err != nil {
					log.Error().Err(err).Msg("failed to set header")
					return nil, status.Errorf(codes.Internal, "failed to set header: %v", err)
				}
			}
		}

		return resp, nil
	}
}
