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
	// ServerAddress address where the server will be hosted (e.g., "localhost:8080" for localhost on port 8080).
	ServerAddress string

	// FileStoragePath path to the file where metrics will be stored on the disk.
	FileStoragePath string

	// DatabaseDsn Data Source Name for the database connection string.
	DatabaseDsn string

	// Key for hashing.
	Key string

	// StoreInterval interval (in seconds) between saving metrics to file.
	StoreInterval int

	// Restore metrics from file after startup or not.
	Restore bool
}

type envs struct {
	address         string `env:"ADDRESS"`
	fileStoragePath string `env:"FILE_STORAGE_PATH"`
	databaseDsn     string `env:"DATABASE_DSN"`
	key             string `env:"KEY"`
	storeInterval   int    `env:"STORE_INTERVAL"`
	restore         bool   `env:"RESTORE"`
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
	flag.Parse()

	envVariables := envs{}
	err := env.Parse(&envVariables)
	if err != nil {
		zap.L().Error("Failed to parse environment variables", zap.Error(err))
	}

	_, exists := os.LookupEnv("ADDRESS")
	if exists && !stringutils.IsEmpty(envVariables.address) {
		config.ServerAddress = envVariables.address
	}

	_, exists = os.LookupEnv("STORE_INTERVAL")
	if exists {
		config.StoreInterval = envVariables.storeInterval
	}

	_, exists = os.LookupEnv("FILE_STORAGE_PATH")
	if exists && !stringutils.IsEmpty(envVariables.fileStoragePath) {
		config.FileStoragePath = envVariables.fileStoragePath
	}

	_, exists = os.LookupEnv("RESTORE")
	if exists {
		config.Restore = envVariables.restore
	}

	_, exists = os.LookupEnv("DATABASE_DSN")
	if exists {
		config.DatabaseDsn = envVariables.databaseDsn
	}

	_, exists = os.LookupEnv("KEY")
	if exists {
		config.Key = envVariables.key
	}

	return &config
}
