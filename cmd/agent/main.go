package main

import (
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal"
	"github.com/zavtra-na-rabotu/gometrics/internal/agent/metrics"
)

func main() {
	Configure()

	// FIXME: "http://"...
	client := resty.New().SetBaseURL("http://" + config.serverAddress)
	metricsCollector := metrics.NewCollector()
	metricsSender := metrics.NewSender(client, metricsCollector)

	collectTicker := time.NewTicker(time.Duration(config.pollInterval) * time.Second)
	senderTicker := time.NewTicker(time.Duration(config.reportInterval) * time.Second)

	defer collectTicker.Stop()
	defer senderTicker.Stop()

	go func() {
		for {
			select {
			case <-collectTicker.C:
				metricsCollector.Collect()
			case <-senderTicker.C:
				err := metricsSender.Send()
				if err != nil {
					internal.ErrorLog.Printf("Error sending metrics: %v", err)
					continue
				}
				metricsCollector.ResetPollCounter()
			}
		}
	}()

	select {}
}
