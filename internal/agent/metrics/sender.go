package metrics

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal"
	"log"
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

func (sender *Sender) Send() {
	for name, metric := range sender.metrics.gaugeMetrics {
		sendMetric(sender.client, internal.Gauge, name, metric)
	}
	for name, metric := range sender.metrics.counterMetrics {
		sendMetric(sender.client, internal.Counter, name, metric)
	}
}

func sendMetric[T int64 | float64](client *resty.Client, metricType internal.MetricType, metricName string, value T) {
	url := fmt.Sprintf("/update/%s/%s/%v", metricType, metricName, value)

	_, err := client.R().Post(url)
	if err != nil {
		log.Printf("Failed to send metric %s: %v", metricName, err)
	}
}
