package metrics

import (
	"bytes"
	"compress/gzip"
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
}

func NewSender(client *resty.Client, metrics *Collector) *Sender {
	return &Sender{
		client:  client,
		metrics: metrics,
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

	return sendMetrics(sender.client, metrics)
}

func sendMetrics(client *resty.Client, metrics []model.Metrics) error {
	var compressedBody bytes.Buffer
	err := compressBody(&compressedBody, metrics)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}

	response, err := client.R().
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

func compressBody(compressedData *bytes.Buffer, metrics []model.Metrics) error {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	gzipWriter := gzip.NewWriter(compressedData)
	_, err = gzipWriter.Write(jsonData)
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
