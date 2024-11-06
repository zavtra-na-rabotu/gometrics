// Package configuration is a package for agent configuration
package configuration

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

// Configuration structure to configure agent parameters
type Configuration struct {
	// ServerAddress address where metrics should be sent (e.g., "localhost:8080" for localhost on port 8080).
	ServerAddress string

	// Key for hashing.
	Key string

	// ReportInterval interval (in seconds) between sending metrics to server
	ReportInterval int

	// PollInterval interval (in seconds) between collecting metrics
	PollInterval int

	// RateLimit limit outgoing requests with metrics
	RateLimit int
}

type envs struct {
	serverAddress  string `env:"ADDRESS"`
	key            string `env:"KEY"`
	reportInterval int    `env:"REPORT_INTERVAL"`
	pollInterval   int    `env:"POLL_INTERVAL"`
	rateLimit      int    `env:"RATE_LIMIT"`
}

// Configure read env variables and CLI parameters to configure server
func Configure() *Configuration {
	var config Configuration

	const defaultReportInterval = 10
	const defaultPollInterval = 2

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Server URL")
	flag.IntVar(&config.ReportInterval, "r", defaultReportInterval, "Report interval in seconds")
	flag.IntVar(&config.PollInterval, "p", defaultPollInterval, "Poll interval in seconds")
	flag.StringVar(&config.Key, "k", "", "Key")
	flag.IntVar(&config.RateLimit, "l", 1, "Rate limit")
	flag.Parse()

	envVariables := envs{}
	err := env.Parse(&envVariables)
	if err != nil {
		zap.L().Error("Failed to parse environment variables", zap.Error(err))
	}

	_, exists := os.LookupEnv("ADDRESS")
	if exists && envVariables.serverAddress != "" {
		config.ServerAddress = envVariables.serverAddress
	}

	_, exists = os.LookupEnv("REPORT_INTERVAL")
	if exists && envVariables.reportInterval != 0 {
		config.ReportInterval = envVariables.reportInterval
	}

	_, exists = os.LookupEnv("POLL_INTERVAL")
	if exists && envVariables.pollInterval != 0 {
		config.PollInterval = envVariables.pollInterval
	}

	_, exists = os.LookupEnv("KEY")
	if exists {
		config.Key = envVariables.key
	}

	_, exists = os.LookupEnv("RATE_LIMIT")
	if exists && envVariables.rateLimit != 0 {
		config.RateLimit = envVariables.rateLimit
	}

	return &config
}
