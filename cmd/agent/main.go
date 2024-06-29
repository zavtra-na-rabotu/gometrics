package main

import (
	"github.com/zavtra-na-rabotu/gometrics/internal/agent/metrics"
	"github.com/zavtra-na-rabotu/gometrics/internal/logger"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
)

func main() {
	logger.InitLogger()
	Configure()

	metricsChan := make(chan []model.Metrics)
	defer close(metricsChan)

	metricsCollector := metrics.NewCollector(config.pollInterval, metricsChan)
	go metricsCollector.InitCollector()
	go metricsCollector.InitPsutilCollector()

	metricsSender := metrics.NewSender(config.serverAddress, config.key, config.rateLimit, config.reportInterval, metricsCollector)
	go metricsSender.InitSender()

	select {}
}
