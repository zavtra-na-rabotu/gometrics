package interceptor

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func IPValidationInterceptor(trustedSubnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Read metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			zap.L().Error("Metadata is missing in the request")
			return nil, status.Errorf(codes.PermissionDenied, "Metadata is missing")
		}

		ipValues := md.Get("X-Real-IP")
		if len(ipValues) == 0 {
			zap.L().Error("X-Real-IP metadata is missing")
			return nil, status.Errorf(codes.PermissionDenied, "X-Real-IP metadata is missing")
		}

		// Parse IP
		realIP := ipValues[0]
		ip := net.ParseIP(realIP)
		if ip == nil {
			zap.L().Error("Invalid IP address in X-Real-IP metadata", zap.String("X-Real-IP", realIP))
			return nil, status.Errorf(codes.PermissionDenied, "Invalid IP address in X-Real-IP metadata")
		}

		// Check if IP in trusted subnet
		if !trustedSubnet.Contains(ip) {
			zap.L().Error("Forbidden: IP not in trusted subnet", zap.String("subnet", trustedSubnet.String()), zap.String("ip", ip.String()))
			return nil, status.Errorf(codes.PermissionDenied, "Forbidden: IP not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
