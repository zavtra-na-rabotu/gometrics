package interceptor

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func HashInterceptor(key string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Read metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			zap.L().Error("Metadata is missing in the request")
			return nil, status.Errorf(codes.PermissionDenied, "Metadata is missing")
		}

		// Read the incoming hash from metadata
		receivedHashes := md.Get("HashSHA256")
		if len(receivedHashes) == 0 {
			zap.L().Error("HashSHA256 metadata is missing")
			return nil, status.Errorf(codes.PermissionDenied, "HashSHA256 metadata is missing")
		}

		receivedHash := receivedHashes[0]

		// Convert request to proto.Message
		protoRequest, ok := req.(proto.Message)
		if !ok {
			zap.L().Error("Request is not proto.Message", zap.Any("Request", req))
			return nil, status.Errorf(codes.Internal, "failed to cast request to proto.Message")
		}

		// Serialize proto request message
		requestBytes, err := proto.Marshal(protoRequest)
		if err != nil {
			zap.L().Error("failed to marshal protobuf request", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to marshal request")
		}

		// Compare received and calculated hash
		calculatedHash := calculateHash(requestBytes, key)
		if receivedHash != "" && receivedHash != calculatedHash {
			zap.L().Error("Hash mismatch", zap.String("received hash", receivedHash), zap.String("calculated hash", calculatedHash))
			return nil, status.Errorf(codes.InvalidArgument, "Hash mismatch")
		}

		// Proceed with the request
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		// Convert response to proto.Message
		protoResponse, ok := resp.(proto.Message)
		if !ok {
			zap.L().Error("Response is not proto.Message", zap.Any("Response", req))
			return nil, status.Errorf(codes.Internal, "failed to cast response to proto.Message")
		}

		// Serialize response to calculate hash
		respBytes, err := proto.Marshal(protoResponse)
		if err != nil {
			zap.L().Error("Failed to serialize response", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "Failed to serialize response")
		}

		responseHash := calculateHash(respBytes, key)

		// Attach the response hash to the outgoing context
		if err := grpc.SetHeader(ctx, metadata.Pairs("HashSHA256", responseHash)); err != nil {
			zap.L().Error("Failed to set response hash", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "Failed to set response hash")
		}

		return resp, nil
	}
}

func calculateHash(data []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
