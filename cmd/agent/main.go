package main

import (
	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/agent/metrics"
	"time"
)

const (
	host            = "http://localhost:8080"
	collectInterval = 2 * time.Second
	sendInterval    = 10 * time.Second
)

func main() {
	client := resty.New().SetBaseURL(host)
	metricsCollector := metrics.NewCollector()
	metricsSender := metrics.NewSender(client, metricsCollector)

	collectTicker := time.NewTicker(collectInterval)
	senderTicker := time.NewTicker(sendInterval)

	defer collectTicker.Stop()
	defer senderTicker.Stop()

	go func() {
		for {
			select {
			case <-collectTicker.C:
				metricsCollector.Collect()
			case <-senderTicker.C:
				metricsSender.Send()
			}
		}
	}()

	select {}
}
