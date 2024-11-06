package main

import (
	"fmt"

	"github.com/zavtra-na-rabotu/gometrics/internal/agent/configuration"
	"github.com/zavtra-na-rabotu/gometrics/internal/agent/metrics"
	"github.com/zavtra-na-rabotu/gometrics/internal/logger"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

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
