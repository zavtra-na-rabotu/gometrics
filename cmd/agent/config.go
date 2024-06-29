package main

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

var config struct {
	serverAddress  string
	key            string
	reportInterval int
	pollInterval   int
	rateLimit      int
}

type envs struct {
	ServerAddress  string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

// Configure TODO: Move to internal.
func Configure() {
	const defaultReportInterval = 10
	const defaultPollInterval = 2

	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "Server URL")
	flag.IntVar(&config.reportInterval, "r", defaultReportInterval, "Report interval in seconds")
	flag.IntVar(&config.pollInterval, "p", defaultPollInterval, "Poll interval in seconds")
	flag.StringVar(&config.key, "k", "", "Key")
	flag.IntVar(&config.rateLimit, "l", 1, "Rate limit")
	flag.Parse()

	envVariables := envs{}
	err := env.Parse(&envVariables)
	if err != nil {
		zap.L().Error("Failed to parse environment variables", zap.Error(err))
	}

	_, exists := os.LookupEnv("ADDRESS")
	if exists && envVariables.ServerAddress != "" {
		config.serverAddress = envVariables.ServerAddress
	}

	_, exists = os.LookupEnv("REPORT_INTERVAL")
	if exists && envVariables.ReportInterval != 0 {
		config.reportInterval = envVariables.ReportInterval
	}

	_, exists = os.LookupEnv("POLL_INTERVAL")
	if exists && envVariables.PollInterval != 0 {
		config.pollInterval = envVariables.PollInterval
	}

	_, exists = os.LookupEnv("KEY")
	if exists {
		config.key = envVariables.Key
	}

	_, exists = os.LookupEnv("RATE_LIMIT")
	if exists && envVariables.RateLimit != 0 {
		config.rateLimit = envVariables.RateLimit
	}
}
