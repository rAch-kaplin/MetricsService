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

// WithHashing returns a gRPC unary interceptor that verifies and adds
// an HMAC-SHA256 hash for requests and responses.
//
// If the incoming metadata contains a "HashSHA256" header, the interceptor
// decodes it and verifies that it matches the hash of the request body.
// If the hash is invalid, it returns Internal error.
//
// After the handler runs successfully, the interceptor calculates a new
// hash for the response and sets it in the response headers as "HashSHA256".
// If the request is not a proto.Message, the interceptor returns Internal error.
func WithHashing(key []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Get the metadata from the incoming context.
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "failed to get metadata")
		}

		// Get the hash from the metadata.
		hashes := md.Get("HashSHA256")
		if len(hashes) == 0 {
			return handler(ctx, req)
		}
		// decoded it's first hash from hashes slice.
		decoded, err := hex.DecodeString(hashes[0])
		if err != nil {
			log.Error().Err(err).Msg("failed to decode hash")
			return nil, status.Errorf(codes.Internal, "failed to decode hash: %v", err)
		}

		// Check if the request is a proto.Message.
		if msg, ok := req.(proto.Message); ok {
			body, err := proto.Marshal(msg)
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal request")
				return nil, status.Errorf(codes.Internal, "failed to marshal request: %v", err)
			}

			// Check if the hash is valid.
			valid := hash.CheckHash(key, body, decoded)
			if !valid {
				log.Error().Msg("invalid hash message")
				return nil, status.Errorf(codes.Internal, "invalid hash message")
			}
		}
		// Run the handler.
		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		// Check if the response is a proto.Message.
		if msg, ok := resp.(proto.Message); ok {
			body, err := proto.Marshal(msg)
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal response")
				return nil, status.Errorf(codes.Internal, "failed to marshal response: %v", err)
			}
			// If the response is not empty, calculate a new hash for the response.
			if len(body) > 0 {
				newHash, err := hash.GetHash(key, body)
				if err != nil {
					log.Error().Err(err).Msg("failed to get new hash")
					return nil, status.Errorf(codes.Internal, "failed to get new hash: %v", err)
				}
				// Set the new hash in the response headers.
				md := metadata.Pairs("HashSHA256", hex.EncodeToString(newHash))
				if err := grpc.SetHeader(ctx, md); err != nil {
					log.Error().Err(err).Msg("failed to set header")
					return nil, status.Errorf(codes.Internal, "failed to set header: %v", err)
				}
			}
		}
		// Return the response.
		return resp, nil
	}
}
