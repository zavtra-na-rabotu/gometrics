package configuration

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"go.uber.org/zap"
)

type Configuration struct {
	ServerAddress   string
	FileStoragePath string
	DatabaseDsn     string
	Key             string
	StoreInterval   int
	Restore         bool
}

type envs struct {
	Address         string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	Restore         bool   `env:"RESTORE"`
}

// Configure TODO: Move to internal.
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

	return &config
}
