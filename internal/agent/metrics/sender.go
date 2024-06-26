package metrics

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"go.uber.org/zap"
)

type Sender struct {
	client  *resty.Client
	metrics *Collector
	key     string
}

func NewSender(client *resty.Client, metrics *Collector, key string) *Sender {
	return &Sender{
		client:  client,
		metrics: metrics,
		key:     key,
	}
}

func (sender *Sender) Send() error {
	var metrics []model.Metrics

	for name, metric := range sender.metrics.gaugeMetrics {
		var gaugeMetric = model.Metrics{ID: name, MType: string(model.Gauge), Value: &metric}
		metrics = append(metrics, gaugeMetric)
	}
	for name, metric := range sender.metrics.counterMetrics {
		var counterMetric = model.Metrics{ID: name, MType: string(model.Counter), Delta: &metric}
		metrics = append(metrics, counterMetric)
	}

	return sendMetrics(sender.client, metrics, sender.key)
}

func sendMetrics(client *resty.Client, metrics []model.Metrics, key string) error {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	var compressedBody bytes.Buffer
	err = compressBody(&compressedBody, jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}

	request := client.R()

	if key != "" {
		var hash = calculateHash(jsonData, key)
		request.SetHeader("HashSHA256", hash)
	}

	response, err := request.
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(compressedBody.Bytes()).
		Post("/updates/")

	if err != nil {
		return fmt.Errorf("failed to send metric %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send metric, StatusCode: %d", response.StatusCode())
	}
	return nil
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

func calculateHash(jsonData []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(jsonData)
	return hex.EncodeToString(hash.Sum(nil))
}
