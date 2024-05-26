package metrics

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
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
	response, err := client.R().
		SetBody(model.Metrics{ID: metricName, MType: string(metricType), Delta: &delta}).
		Post("/update/")

	if err != nil {
		return fmt.Errorf("failed to send counter metric %s: %w", metricName, err)
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send counter metric %s, StatusCode: %d", metricName, response.StatusCode())
	}
	return nil
}

func sendGaugeMetric(client *resty.Client, metricType model.MetricType, metricName string, value float64) error {
	response, err := client.R().
		SetBody(model.Metrics{ID: metricName, MType: string(metricType), Value: &value}).
		Post("/update/")

	if err != nil {
		return fmt.Errorf("failed to send gauge metric %s: %w", metricName, err)
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send gauge metric %s, StatusCode: %d", metricName, response.StatusCode())
	}
	return nil
}
