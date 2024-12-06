package v1

import (
	"context"
	"encoding/json"

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
	var metrics []model.Metrics

	err := json.Unmarshal(req.Data, &metrics)
	if err != nil {
		zap.L().Error("Error unmarshalling data", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
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
