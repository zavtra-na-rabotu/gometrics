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
		err := sendMetric(sender.client, model.Gauge, name, metric)
		if err != nil {
			return err
		}
	}
	for name, metric := range sender.metrics.counterMetrics {
		err := sendMetric(sender.client, model.Counter, name, metric)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendMetric[T int64 | float64](
	client *resty.Client,
	metricType model.MetricType,
	metricName string,
	value T,
) error {
	url := fmt.Sprintf("/update/%s/%s/%v", metricType, metricName, value)

	response, err := client.R().Post(url)
	if err != nil {
		return fmt.Errorf("failed to send metric %s: %w", metricName, err)
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send metric %s, StatusCode: %d", metricName, response.StatusCode())
	}
	return nil
}
