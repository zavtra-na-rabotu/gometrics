package main

import (
	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/agent/metrics"
	"time"
)

func main() {
	Configure()

	// FIXME: "http://"...
	client := resty.New().SetBaseURL("http://" + serverAddress)
	metricsCollector := metrics.NewCollector()
	metricsSender := metrics.NewSender(client, metricsCollector)

	collectTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	senderTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)

	defer collectTicker.Stop()
	defer senderTicker.Stop()

	go func() {
		for {
			select {
			case <-collectTicker.C:
				metricsCollector.Collect()
			case <-senderTicker.C:
				metricsSender.Send()
				metricsCollector.ResetPollCounter()
			}
		}
	}()

	select {}
}
