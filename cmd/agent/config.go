package main

import (
	"flag"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

var config struct {
	serverAddress  string
	reportInterval int
	pollInterval   int
}

type envs struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func Configure() {
	const defaultReportInterval = 10
	const defaultPollInterval = 2

	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "Server URL")
	flag.IntVar(&config.reportInterval, "r", defaultReportInterval, "Report interval in seconds")
	flag.IntVar(&config.pollInterval, "p", defaultPollInterval, "Poll interval in seconds")
	flag.Parse()

	envVariables := envs{}
	err := env.Parse(&envVariables)
	if err != nil {
		zap.L().Error("Failed to parse environment variables", zap.Error(err))
	}

	if envVariables.ServerAddress != "" {
		config.serverAddress = envVariables.ServerAddress
	}
	if envVariables.ReportInterval != 0 {
		config.reportInterval = envVariables.ReportInterval
	}
	if envVariables.PollInterval != 0 {
		config.pollInterval = envVariables.PollInterval
	}
}
