package metrics

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"go.uber.org/zap"
)

type HTTPClient struct {
	client    *resty.Client
	key       string
	publicKey *rsa.PublicKey
	localIP   net.IP
}

func NewHTTPClient(serverAddress string, key string, publicKey *rsa.PublicKey) *HTTPClient {
	return &HTTPClient{
		client:    resty.New().SetBaseURL("http://" + serverAddress),
		key:       key,
		publicKey: publicKey,
		localIP:   GetLocalIP(),
	}
}

// SendMetrics HTTP sender implementation
func (httpClient *HTTPClient) SendMetrics(metrics []model.Metrics) error {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics:  %w", err)
	}

	var compressedBody bytes.Buffer
	err = compressBody(&compressedBody, jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}

	var encryptedAESKey string
	var encryptedData = compressedBody.Bytes()
	if httpClient.publicKey != nil {
		encryptedData, encryptedAESKey, err = encryptRequestBody(compressedBody.Bytes(), httpClient.publicKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt metrics: %w", err)
		}
	}

	request := httpClient.client.R()

	if httpClient.key != "" {
		var hash = CalculateHash(jsonData, httpClient.key)
		request.SetHeader("HashSHA256", hash)
	}

	response, err := request.
		SetHeader("X-Real-IP", httpClient.localIP.String()).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetHeader("Encrypted-AES-Key", encryptedAESKey).
		SetBody(encryptedData).
		Post("/updates/")

	if err != nil {
		return fmt.Errorf("failed to send metric %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send metric, StatusCode: %d", response.StatusCode())
	}
	return nil
}

// encryptWithPublicKey encrypts data using RSA public key
func encryptWithPublicKey(data []byte, publicKey *rsa.PublicKey) (string, error) {
	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// encryptWithAES encrypts data using AES-GCM
func encryptWithAES(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	encryptedData := aesGCM.Seal(nonce, nonce, data, nil)
	return encryptedData, nil
}

func encryptRequestBody(body []byte, publicKey *rsa.PublicKey) ([]byte, string, error) {
	// Generate AES key
	aesKey := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, aesKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Encrypt data with AES key
	encryptedData, err := encryptWithAES(body, aesKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt data with AES: %w", err)
	}

	// Encrypt AES key with RSA public key
	encryptedAESKey, err := encryptWithPublicKey(aesKey, publicKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	return encryptedData, encryptedAESKey, nil
}

func compressBody(compressedData *bytes.Buffer, jsonData []byte) error {
	gzipWriter := gzip.NewWriter(compressedData)
	_, err := gzipWriter.Write(jsonData)
	if err != nil {
		zap.L().Error("Error compressing data", zap.Error(err))
		return fmt.Errorf("error compressing data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		zap.L().Error("Error closing gzip writer", zap.Error(err))
		return fmt.Errorf("error closing gzip writer: %w", err)
	}

	return nil
}
