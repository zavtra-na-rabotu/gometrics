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
	for name, metric := range sender.metrics.gaugeMetrics {
		err := sendGaugeMetric(sender.client, model.Gauge, name, metric)
		if err != nil {
			return err
		}
	}
	for name, metric := range sender.metrics.counterMetrics {
		err := sendCounterMetric(sender.client, model.Counter, name, metric)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendCounterMetric(client *resty.Client, metricType model.MetricType, metricName string, delta int64) error {
	var compressedBody bytes.Buffer
	var metrics = model.Metrics{ID: metricName, MType: string(metricType), Delta: &delta}
	err := compressBody(&compressedBody, metrics)
	if err != nil {
		return fmt.Errorf("failed to compress counter metric %s: %w", metricName, err)
	}

	err = sendMetric(client, &compressedBody)
	if err != nil {
		return fmt.Errorf("failed to send counter metric %s: %w", metricName, err)
	}

	return nil
}

func sendGaugeMetric(client *resty.Client, metricType model.MetricType, metricName string, value float64) error {
	var compressedBody bytes.Buffer
	var metrics = model.Metrics{ID: metricName, MType: string(metricType), Value: &value}
	err := compressBody(&compressedBody, metrics)
	if err != nil {
		return fmt.Errorf("failed to compress gauge metric %s: %w", metricName, err)
	}

	err = sendMetric(client, &compressedBody)
	if err != nil {
		return fmt.Errorf("failed to send gauge metric %s: %w", metricName, err)
	}

	return nil
}

func sendMetric(client *resty.Client, compressedBody *bytes.Buffer) error {
	response, err := client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(compressedBody.Bytes()).
		Post("/update/")

	if err != nil {
		return fmt.Errorf("failed to send metric %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send metric, StatusCode: %d", response.StatusCode())
	}
	return nil
}

func compressBody(compressedData *bytes.Buffer, metrics model.Metrics) error {
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
