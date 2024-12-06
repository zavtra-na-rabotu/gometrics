package interceptor

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/zavtra-na-rabotu/gometrics/internal/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func DecryptionInterceptor(privateKey *rsa.PrivateKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Read metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			zap.L().Error("Metadata is missing in the request")
			return nil, status.Errorf(codes.InvalidArgument, "Metadata is missing")
		}

		// Get encrypted AES from metadata
		encryptedAESKeyMetadata := md.Get("Encrypted-AES-Key")
		if len(encryptedAESKeyMetadata) == 0 {
			zap.L().Error("AES key metadata is missing")
			return nil, status.Errorf(codes.InvalidArgument, "AES key metadata is missing")
		}

		encryptedAESKey := encryptedAESKeyMetadata[0]

		// Decrypt AES key
		aesKey, err := decryptWithPrivateKey(encryptedAESKey, privateKey)
		if err != nil {
			zap.L().Error("Failed to decrypt AES key", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "Failed to decrypt AES key")
		}

		// Check if the request is of type UpdateMetricsRequest
		request, ok := req.(*pb.UpdateMetricsRequest)
		if !ok {
			zap.L().Error("Request is not of type UpdateMetricsRequest")
			return nil, status.Errorf(codes.Internal, "Request is not of type UpdateMetricsRequest")
		}

		// Decrypt data
		decryptedData, err := decryptWithAES(request.Data, aesKey)
		if err != nil {
			zap.L().Error("Failed to decrypt request data", zap.Error(err))
			return nil, fmt.Errorf("failed to decrypt request data: %w", err)
		}

		// Replace the encrypted data with decrypted data
		request.Data = decryptedData

		return handler(ctx, request)
	}
}

// decryptWithPrivateKey decrypt AES key using RSA private key
func decryptWithPrivateKey(encryptedKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted AES key: %w", err)
	}

	aesKey, err := rsa.DecryptOAEP(sha256.New(), nil, privateKey, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	return aesKey, nil
}

// decryptWithAES decrypt request body using AES key
func decryptWithAES(encryptedData []byte, aesKey []byte) ([]byte, error) {
	nonceSize := 12
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data is too short")
	}

	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
