package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zavtra-na-rabotu/gometrics/internal/logger"
	v1 "github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/v1"
	v2 "github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/middleware"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	Configure()

	memStorage := storage.NewMemStorage()

	r := chi.NewRouter()
	r.Use(middleware.RequestLoggerMiddleware)
	r.Use(middleware.GzipMiddleware)

	// API v1
	r.Post("/update/{type}/{name}/{value}", v1.UpdateMetric(memStorage))
	r.Get("/value/{type}/{name}", v1.GetMetric(memStorage))
	r.Get("/", v1.RenderAllMetrics(memStorage))

	// API v2
	r.Post("/update/", v2.UpdateMetric(memStorage))
	r.Post("/value/", v2.GetMetric(memStorage))

	err := storage.ConfigureStorage(memStorage, config.fileStoragePath, config.restore, config.storeInterval)
	if err != nil {
		zap.L().Fatal("failed to configure storage", zap.Error(err))
	}

	err = http.ListenAndServe(config.serverAddress, r)
	if err != nil {
		zap.L().Fatal("Failed to start server", zap.Error(err))
	}
}
