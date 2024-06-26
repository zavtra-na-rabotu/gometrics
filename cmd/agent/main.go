package main

import (
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/agent/metrics"
	"github.com/zavtra-na-rabotu/gometrics/internal/logger"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	Configure()

	// FIXME: "http://"...
	client := resty.New().SetBaseURL("http://" + config.serverAddress)
	metricsCollector := metrics.NewCollector()
	metricsSender := metrics.NewSender(client, metricsCollector, config.key)

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
					zap.L().Error("Failed to send metrics to server", zap.Error(err))
					continue
				}
				metricsCollector.ResetPollCounter()
			}
		}
	}()

	select {}
}
