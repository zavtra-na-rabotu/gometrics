// Package configuration read env variables and CLI parameters to configure server
package configuration

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"go.uber.org/zap"
)

// Configuration structure to configure server parameters
type Configuration struct {
	// ServerAddress Address where the server will be hosted (e.g., "localhost:8080" for localhost on port 8080).
	ServerAddress string

	// FileStoragePath path to the file where metrics will be stored on the disk.
	FileStoragePath string

	// DatabaseDsn Data Source Name for the database connection string.
	DatabaseDsn string

	// CryptoKey path to private Key for request decryption
	CryptoKey string

	// Key for hashing.
	Key string

	// StoreInterval interval (in seconds) between saving metrics to file.
	StoreInterval int

	// Restore metrics from file after startup or not.
	Restore bool
}

type envs struct {
	Address         string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	CryptoKey       string `env:"CRYPTO_KEY"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	Restore         bool   `env:"RESTORE"`
}

// Configure read env variables and CLI parameters to configure server
func Configure() *Configuration {
	var config Configuration

	const defaultStoreInterval = 300
	const defaultFileStoragePath = "/tmp/metrics-db.json"
	const defaultRestore = true

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Server URL")
	flag.IntVar(&config.StoreInterval, "i", defaultStoreInterval, "Store interval in seconds")
	flag.StringVar(&config.FileStoragePath, "f", defaultFileStoragePath, "File storage path")
	flag.BoolVar(&config.Restore, "r", defaultRestore, "Restore")
	flag.StringVar(&config.DatabaseDsn, "d", "", "Database DSN")
	flag.StringVar(&config.Key, "k", "", "Key")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "Crypto Key")
	flag.Parse()

	var envVariables envs
	err := env.Parse(&envVariables)
	if err != nil {
		zap.L().Error("Failed to parse environment variables", zap.Error(err))
	}

	_, exists := os.LookupEnv("ADDRESS")
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

	return &config
}
