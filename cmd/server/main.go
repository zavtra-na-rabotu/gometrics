package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zavtra-na-rabotu/gometrics/internal/logger"
	v1 "github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/v1"
	v2 "github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/v2"
	v3 "github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/v3"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/middleware"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	Configure()

	r := chi.NewRouter()
	r.Use(middleware.RequestLoggerMiddleware)
	r.Use(middleware.GzipMiddleware)

	var storageToUse storage.Storage

	// TODO: Выглядит плохо, но я пока не знаю как лучше
	if !stringutils.IsEmpty(config.databaseDsn) {
		zap.L().Info("Using database storage")

		dbStorage, err := storage.NewDBStorage(config.databaseDsn)
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

		err := storage.ConfigureStorage(storageToUse.(*storage.MemStorage), config.fileStoragePath, config.restore, config.storeInterval)
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

	err := http.ListenAndServe(config.serverAddress, r)
	if err != nil {
		zap.L().Fatal("Failed to start server", zap.Error(err))
	}
}
