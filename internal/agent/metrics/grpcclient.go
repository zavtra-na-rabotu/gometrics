package metrics

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
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
	publicKey     *rsa.PublicKey
	localIP       net.IP
}

func NewGRPCClient(serverAddress string, key string, publicKey *rsa.PublicKey) *GRPCClient {
	return &GRPCClient{serverAddress, key, publicKey, getLocalIP()}
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

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		zap.L().Error("Error marshalling metrics", zap.Error(err))
		return fmt.Errorf("failed to marshal metrics:  %w", err)
	}

	var compressedBody bytes.Buffer
	err = compressBody(&compressedBody, jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}

	var encryptedAESKey string
	var encryptedData = compressedBody.Bytes()
	if grpcClient.publicKey != nil {
		encryptedData, encryptedAESKey, err = encryptRequestBody(compressedBody.Bytes(), grpcClient.publicKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt metrics: %w", err)
		}
	}

	// Create request
	request := &pb.UpdateMetricsRequest{
		Data: encryptedData,
	}

	// Create metadata
	md := metadata.New(map[string]string{
		"X-Real-IP":         grpcClient.localIP.String(),
		"Encrypted-AES-Key": encryptedAESKey,
		"Content-Encoding":  "gzip",
	})

	// If key exist - calculate hash and append to metadata
	if grpcClient.key != "" {
		requestBytes, err := proto.Marshal(request)
		if err != nil {
			zap.L().Error("Error marshaling request", zap.Error(err))
			return err
		}

		var hash = calculateHash(requestBytes, grpcClient.key)
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
