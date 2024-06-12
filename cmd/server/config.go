package main

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"go.uber.org/zap"
)

var config struct {
	serverAddress   string
	fileStoragePath string
	storeInterval   int
	restore         bool
	databaseDsn     string
}

type envs struct {
	Address         string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
}

// Configure TODO: Move to internal.
func Configure() {
	const defaultStoreInterval = 300
	const defaultFileStoragePath = "/tmp/metrics-db.json"
	const defaultRestore = true

	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "Server URL")
	flag.IntVar(&config.storeInterval, "i", defaultStoreInterval, "Store interval in seconds")
	flag.StringVar(&config.fileStoragePath, "f", defaultFileStoragePath, "File storage path")
	flag.BoolVar(&config.restore, "r", defaultRestore, "Restore")
	flag.StringVar(&config.databaseDsn, "d", "", "Database DSN")
	flag.Parse()

	envVariables := envs{}
	err := env.Parse(&envVariables)
	if err != nil {
		zap.L().Error("Failed to parse environment variables", zap.Error(err))
	}

	_, exists := os.LookupEnv("ADDRESS")
	if exists && !stringutils.IsEmpty(envVariables.Address) {
		config.serverAddress = envVariables.Address
	}

	_, exists = os.LookupEnv("STORE_INTERVAL")
	if exists {
		config.storeInterval = envVariables.StoreInterval
	}

	_, exists = os.LookupEnv("FILE_STORAGE_PATH")
	if exists && !stringutils.IsEmpty(envVariables.FileStoragePath) {
		config.fileStoragePath = envVariables.FileStoragePath
	}

	_, exists = os.LookupEnv("RESTORE")
	if exists {
		config.restore = envVariables.Restore
	}

	_, exists = os.LookupEnv("DATABASE_DSN")
	if exists {
		config.databaseDsn = envVariables.DatabaseDsn
	}
}
