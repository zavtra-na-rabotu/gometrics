package metrics

import (
	"context"
	"net"
	"time"

	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"github.com/zavtra-na-rabotu/gometrics/internal/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type GRPCClient struct {
	serverAddress string
	key           string
	localIP       net.IP
}

func NewGRPCClient(serverAddress string, key string) *GRPCClient {
	return &GRPCClient{serverAddress, key, GetLocalIP()}
}

// SendMetrics gRPC sender implementation
func (grpcClient *GRPCClient) SendMetrics(metrics []model.Metrics) error {
	connection, err := grpc.NewClient(grpcClient.serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.L().Error("Error creating client", zap.Error(err))
		return err
	}
	defer connection.Close()

	client := pb.NewMetricsServiceClient(connection)

	// Create request
	request := &pb.UpdateMetricsRequest{
		Metrics: mapModelMetricsToPb(metrics),
	}

	// Create metadata
	md := metadata.New(map[string]string{
		"X-Real-IP": grpcClient.localIP.String(),
	})

	// If key exist - calculate hash and append to metadata
	if grpcClient.key != "" {
		requestBytes, err := proto.Marshal(request)
		if err != nil {
			zap.L().Error("Error marshaling request", zap.Error(err))
			return err
		}

		var hash = CalculateHash(requestBytes, grpcClient.key)
		md.Append("HashSHA256", hash)
	}

	requestContext, requestCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer requestCancel()

	requestContext = metadata.NewOutgoingContext(requestContext, md)

	response, err := client.UpdateMetrics(requestContext, request)
	if err != nil {
		zap.L().Error("Error in sending metrics", zap.Error(err))
		return err
	}

	zap.L().Info("Successfully updated metrics", zap.Reflect("response", response))

	return nil
}

func mapModelMetricsToPb(metrics []model.Metrics) []*pb.Metrics {
	result := make([]*pb.Metrics, len(metrics))
	for i, m := range metrics {
		result[i] = &pb.Metrics{
			Id:    m.ID,
			Type:  mapProtobufTypeToString(m.MType),
			Value: float64PointerToPb(m.Value),
			Delta: int64PointerToPb(m.Delta),
		}
	}
	return result
}

func mapProtobufTypeToString(t string) pb.Metrics_Type {
	if t == "counter" {
		return pb.Metrics_COUNTER
	}
	return pb.Metrics_GAUGE
}

func float64PointerToPb(value *float64) *float64 {
	if value != nil {
		return value
	}
	return nil
}

func int64PointerToPb(delta *int64) *int64 {
	if delta != nil {
		return delta
	}
	return nil
}
