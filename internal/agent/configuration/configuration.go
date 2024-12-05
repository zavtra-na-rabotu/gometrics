// Package configuration is a package for agent configuration
package configuration

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

// Configuration structure to configure agent parameters
type Configuration struct {
	// ServerAddress address where metrics should be sent (e.g., "localhost:8080" for localhost on port 8080).
	ServerAddress string `json:"server_address"`

	// Key for hashing.
	Key string `json:"key"`

	// CryptoKey path to public Key for request encryption
	CryptoKey string `json:"crypto_key"`

	// Config path to configuration file
	Config string `json:"config"`

	// ReportInterval interval (in seconds) between sending metrics to server
	ReportInterval int `json:"report_interval"`

	// PollInterval interval (in seconds) between collecting metrics
	PollInterval int `json:"poll_interval"`

	// RateLimit limit outgoing requests with metrics
	RateLimit int `json:"rate_limit"`

	// GRPCEnabled to enable GRPc instead of HTTP
	GRPCEnabled bool `json:"grpc_enabled"`
}

type envs struct {
	ServerAddress  string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	Config         string `env:"CONFIG"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
	GRPCEnabled    bool   `env:"GRPC_ENABLED"`
}

// Configure read env variables and CLI parameters to configure server
func Configure() *Configuration {
	var config Configuration

	const defaultReportInterval = 10
	const defaultPollInterval = 2

	// Flags for config
	flag.StringVar(&config.Config, "c", "", "Path to configuration file")
	flag.StringVar(&config.Config, "config", "", "Path to configuration file")

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Server URL")
	flag.IntVar(&config.ReportInterval, "r", defaultReportInterval, "Report interval in seconds")
	flag.IntVar(&config.PollInterval, "p", defaultPollInterval, "Poll interval in seconds")
	flag.StringVar(&config.Key, "k", "", "Key")
	flag.IntVar(&config.RateLimit, "l", 1, "Rate limit")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "Crypto Key")
	flag.BoolVar(&config.GRPCEnabled, "g", false, "GRPC enabled")
	flag.Parse()

	var envVariables envs
	err := env.Parse(&envVariables)
	if err != nil {
		zap.L().Error("Failed to parse environment variables", zap.Error(err))
	}

	_, exists := os.LookupEnv("CONFIG")
	if exists && envVariables.Config != "" {
		config.Config = envVariables.Config

		configFromFile, err := loadConfigFromFile(config.Config)
		if err != nil {
			zap.L().Error("Failed to load configuration file", zap.Error(err))
		} else {
			config = *configFromFile
		}
	}

	_, exists = os.LookupEnv("ADDRESS")
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

	_, exists = os.LookupEnv("CRYPTO_KEY")
	if exists && envVariables.CryptoKey != "" {
		config.CryptoKey = envVariables.CryptoKey
	}

	_, exists = os.LookupEnv("GRPC_ENABLED")
	if exists {
		config.GRPCEnabled = envVariables.GRPCEnabled
	}

	return &config
}

// loadConfigFromFile loads configuration from a JSON file
func loadConfigFromFile(filePath string) (*Configuration, error) {
	if filePath == "" {
		return nil, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Configuration
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
