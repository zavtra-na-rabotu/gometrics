package metrics

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
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
		localIP:   getLocalIP(),
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
		var hash = calculateHash(jsonData, httpClient.key)
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
