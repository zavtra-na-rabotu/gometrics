package interceptor

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"strings"

	"github.com/zavtra-na-rabotu/gometrics/internal/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GzipInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Read metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			zap.L().Error("Metadata is missing in the request")
			return nil, status.Errorf(codes.InvalidArgument, "Metadata is missing")
		}

		contentEncodings := md.Get("Content-Encoding")
		if len(contentEncodings) == 0 {
			zap.L().Error("Content-Encoding is missing")
			return nil, status.Errorf(codes.InvalidArgument, "Content-Encoding is missing")
		}

		contentEncoding := contentEncodings[0]

		// Check if the request is of type UpdateMetricsRequest
		request, ok := req.(*pb.UpdateMetricsRequest)
		if !ok {
			zap.L().Error("Request is not of type UpdateMetricsRequest")
			return nil, status.Errorf(codes.Internal, "Request is not of type UpdateMetricsRequest")
		}

		receivedGzip := strings.Contains(contentEncoding, "gzip")
		if receivedGzip {
			reader, err := gzip.NewReader(bytes.NewReader(request.Data))
			if err != nil {
				zap.L().Error("Error creating gzip reader", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "Error creating gzip reader")
			}
			defer reader.Close()

			decompressedData, err := io.ReadAll(reader)
			if err != nil {
				zap.L().Error("Error reading gzipped data", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "Error reading gzipped data")
			}

			request.Data = decompressedData
		}

		return handler(ctx, request)
	}
}
