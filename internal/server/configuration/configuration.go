// Package configuration read env variables and CLI parameters to configure server
package configuration

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"go.uber.org/zap"
)

// Configuration structure to configure server parameters
type Configuration struct {
	// ServerAddress Address where the server will be hosted (e.g., "localhost:8080" for localhost on port 8080).
	ServerAddress string `json:"server_address"`

	// FileStoragePath path to the file where metrics will be stored on the disk.
	FileStoragePath string `json:"file_storage_path"`

	// DatabaseDsn Data Source Name for the database connection string.
	DatabaseDsn string `json:"database_dsn"`

	// CryptoKey path to private Key for request decryption
	CryptoKey string `json:"crypto_key"`

	// Config path to configuration file
	Config string `json:"config"`

	// Key for hashing.
	Key string `json:"key"`

	// TrustedSubnet trusted subnet CIDR
	TrustedSubnet string `json:"trusted_subnet"`

	// StoreInterval interval (in seconds) between saving metrics to file.
	StoreInterval int `json:"store_interval"`

	// Restore metrics from file after startup or not.
	Restore bool `json:"restore"`

	// GRPCEnabled to enable GRPc instead of HTTP
	GRPCEnabled bool `json:"grpc_enabled"`

	// GRPCPort gRPC port to listen
	GRPCPort int `json:"grpc_port"`
}

type envs struct {
	Address         string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	CryptoKey       string `env:"CRYPTO_KEY"`
	Config          string `env:"CONFIG"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	Restore         bool   `env:"RESTORE"`
	GRPCEnabled     bool   `env:"GRPC_ENABLED"`
	GRPCPort        int    `env:"GRPC_PORT"`
}

// Configure read env variables and CLI parameters to configure server
func Configure() *Configuration {
	var config Configuration

	const defaultStoreInterval = 300
	const defaultFileStoragePath = "/tmp/metrics-db.json"
	const defaultRestore = true

	// Flags for config
	flag.StringVar(&config.Config, "c", "", "Path to configuration file")
	flag.StringVar(&config.Config, "config", "", "Path to configuration file")

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Server URL")
	flag.IntVar(&config.StoreInterval, "i", defaultStoreInterval, "Store interval in seconds")
	flag.StringVar(&config.FileStoragePath, "f", defaultFileStoragePath, "File storage path")
	flag.BoolVar(&config.Restore, "r", defaultRestore, "Restore")
	flag.StringVar(&config.DatabaseDsn, "d", "", "Database DSN")
	flag.StringVar(&config.Key, "k", "", "Key")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "Crypto Key")
	flag.StringVar(&config.TrustedSubnet, "t", "", "Trusted subnet CIDR")
	flag.BoolVar(&config.GRPCEnabled, "g", false, "GRPC enabled")
	flag.IntVar(&config.GRPCPort, "p", 50051, "GRPC port")
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
	if exists && !stringutils.IsEmpty(envVariables.Address) {
		config.ServerAddress = envVariables.Address
	}

	_, exists = os.LookupEnv("STORE_INTERVAL")
	if exists {
		config.StoreInterval = envVariables.StoreInterval
	}

	_, exists = os.LookupEnv("FILE_STORAGE_PATH")
	if exists && !stringutils.IsEmpty(envVariables.FileStoragePath) {
		config.FileStoragePath = envVariables.FileStoragePath
	}

	_, exists = os.LookupEnv("RESTORE")
	if exists {
		config.Restore = envVariables.Restore
	}

	_, exists = os.LookupEnv("DATABASE_DSN")
	if exists {
		config.DatabaseDsn = envVariables.DatabaseDsn
	}

	_, exists = os.LookupEnv("KEY")
	if exists {
		config.Key = envVariables.Key
	}

	_, exists = os.LookupEnv("CRYPTO_KEY")
	if exists && envVariables.CryptoKey != "" {
		config.CryptoKey = envVariables.CryptoKey
	}

	_, exists = os.LookupEnv("TRUSTED_SUBNET")
	if exists && envVariables.TrustedSubnet != "" {
		config.TrustedSubnet = envVariables.TrustedSubnet
	}

	_, exists = os.LookupEnv("GRPC_ENABLED")
	if exists {
		config.GRPCEnabled = envVariables.GRPCEnabled
	}

	_, exists = os.LookupEnv("GRPC_PORT")
	if exists && envVariables.GRPCPort != 0 {
		config.GRPCPort = envVariables.GRPCPort
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
