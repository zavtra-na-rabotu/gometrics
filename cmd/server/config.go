package main

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

var config struct {
	serverAddress   string
	storeInterval   int
	fileStoragePath string
	restore         bool
}

type envs struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

// Configure TODO: Move to internal
func Configure() {
	const defaultStoreInterval = 300
	const defaultFileStoragePath = "/tmp/metrics-db.json"
	const defaultRestore = true

	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "Server URL")
	flag.IntVar(&config.storeInterval, "i", defaultStoreInterval, "Store interval in seconds")
	flag.StringVar(&config.fileStoragePath, "f", defaultFileStoragePath, "File storage path")
	flag.BoolVar(&config.restore, "r", defaultRestore, "Restore")
	flag.Parse()

	envVariables := envs{}
	err := env.Parse(&envVariables)
	if err != nil {
		zap.L().Error("Failed to parse environment variables", zap.Error(err))
	}

	_, exists := os.LookupEnv("ADDRESS")
	if exists && envVariables.Address != "" {
		config.serverAddress = envVariables.Address
	}

	_, exists = os.LookupEnv("STORE_INTERVAL")
	if exists && envVariables.StoreInterval != 0 {
		config.storeInterval = envVariables.StoreInterval
	}

	_, exists = os.LookupEnv("FILE_STORAGE_PATH")
	if exists && envVariables.FileStoragePath != "" {
		config.fileStoragePath = envVariables.FileStoragePath
	}

	_, exists = os.LookupEnv("RESTORE")
	if exists {
		config.restore = envVariables.Restore
	}
}
