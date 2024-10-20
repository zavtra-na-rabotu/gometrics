package configuration

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

type Configuration struct {
	ServerAddress  string
	Key            string
	ReportInterval int
	PollInterval   int
	RateLimit      int
}

type envs struct {
	ServerAddress  string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

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
	if exists && envVariables.ServerAddress != "" {
		config.ServerAddress = envVariables.ServerAddress
	}

	_, exists = os.LookupEnv("REPORT_INTERVAL")
	if exists && envVariables.ReportInterval != 0 {
		config.ReportInterval = envVariables.ReportInterval
	}

	_, exists = os.LookupEnv("POLL_INTERVAL")
	if exists && envVariables.PollInterval != 0 {
		config.PollInterval = envVariables.PollInterval
	}

	_, exists = os.LookupEnv("KEY")
	if exists {
		config.Key = envVariables.Key
	}

	_, exists = os.LookupEnv("RATE_LIMIT")
	if exists && envVariables.RateLimit != 0 {
		config.RateLimit = envVariables.RateLimit
	}

	return &config
}
