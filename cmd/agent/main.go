package main

import (
	"github.com/zavtra-na-rabotu/gometrics/internal/agent/configuration"
	"github.com/zavtra-na-rabotu/gometrics/internal/agent/metrics"
	"github.com/zavtra-na-rabotu/gometrics/internal/logger"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
)

func main() {
	config := configuration.Configure()
	logger.InitLogger()

	metricsChan := make(chan []model.Metrics)
	defer close(metricsChan)

	metricsCollector := metrics.NewCollector(config.PollInterval, metricsChan)
	go metricsCollector.InitCollector()
	go metricsCollector.InitPsutilCollector()

	metricsSender := metrics.NewSender(config.ServerAddress, config.Key, config.RateLimit, config.ReportInterval, metricsCollector)
	go metricsSender.InitSender()

	select {}
}
