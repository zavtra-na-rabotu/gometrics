package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	profilermiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/zavtra-na-rabotu/gometrics/internal/logger"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/configuration"
	v1 "github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/v1"
	v2 "github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/v2"
	v3 "github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/v3"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/middleware"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/security"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	config := configuration.Configure()
	logger.InitLogger()

	r := chi.NewRouter()

	if config.CryptoKey != "" {
		privateKey, err := security.ParsePrivateKey(config.CryptoKey)
		if err != nil {
			zap.L().Fatal("Failed to parse crypto key", zap.Error(err))
		}

		r.Use(middleware.DecryptMiddleware(privateKey))
	}

	r.Use(middleware.RequestLoggerMiddleware)
	r.Use(middleware.GzipMiddleware)

	if config.Key != "" {
		r.Use(middleware.RequestHashMiddleware(config.Key))
		r.Use(middleware.ResponseHashMiddleware(config.Key))
	}

	var storageToUse storage.Storage

	if !stringutils.IsEmpty(config.DatabaseDsn) {
		zap.L().Info("Using database storage")

		dbStorage, err := storage.NewDBStorage(config.DatabaseDsn)
		if err != nil {
			zap.L().Fatal("Failed to connect to database", zap.Error(err))
		}
		defer dbStorage.Close()

		err = dbStorage.RunMigrations()
		if err != nil {
			zap.L().Fatal("Failed to run migrations", zap.Error(err))
		}

		storageToUse = dbStorage
	} else {
		zap.L().Info("Using in memory storage")

		storageToUse = storage.NewMemStorage()

		err := storage.ConfigureStorage(storageToUse.(*storage.MemStorage), config.FileStoragePath, config.Restore, config.StoreInterval)
		if err != nil {
			zap.L().Fatal("failed to configure storage", zap.Error(err))
		}
	}

	// API v1
	r.Post("/update/{type}/{name}/{value}", v1.UpdateMetric(storageToUse))
	r.Get("/value/{type}/{name}", v1.GetMetric(storageToUse))
	r.Get("/", v1.RenderAllMetrics(storageToUse))

	// API v2
	r.Post("/update/", v2.UpdateMetric(storageToUse))
	r.Post("/value/", v2.GetMetric(storageToUse))

	// API v3
	r.Get("/ping", v3.Ping(storageToUse))
	r.Post("/updates/", v3.UpdateMetrics(storageToUse))

	// Profiler
	r.Mount("/debug", profilermiddleware.Profiler())

	err := http.ListenAndServe(config.ServerAddress, r)
	if err != nil {
		zap.L().Fatal("Failed to start server", zap.Error(err))
	}
}
