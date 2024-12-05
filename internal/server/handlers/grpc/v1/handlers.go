package v1

import (
	"context"

	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"github.com/zavtra-na-rabotu/gometrics/internal/pb"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedMetricsServiceServer
	storage storage.Storage
}

func NewServer(st storage.Storage) *Server {
	return &Server{storage: st}
}

// UpdateMetrics handler to update batch of metrics
func (s *Server) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	metrics := make([]model.Metrics, len(req.Metrics))
	for i, m := range req.Metrics {
		metrics[i] = model.Metrics{
			ID:    m.Id,
			MType: mapProtobufTypeToString(m.Type),
			Value: optionalFloat64(m.Value),
			Delta: optionalInt64(m.Delta),
		}
	}

	if err := s.storage.UpdateMetrics(metrics); err != nil {
		zap.L().Error("Failed to update metrics", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update metrics")
	}

	return &pb.UpdateMetricsResponse{
		Status:  "ok",
		Message: "Metrics updated successfully",
	}, nil
}

// optionalFloat64 converts *float64 from Protobuf to nil or value
func optionalFloat64(value *float64) *float64 {
	if value != nil {
		return value
	}
	return nil
}

// optionalInt64 converts *int64 from Protobuf to nil or value
func optionalInt64(value *int64) *int64 {
	if value != nil {
		return value
	}
	return nil
}

// mapProtobufTypeToString converts Enum to string
func mapProtobufTypeToString(t pb.Metrics_Type) string {
	if t == pb.Metrics_COUNTER {
		return "counter"
	}
	return "gauge"
}
