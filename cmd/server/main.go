package main

import (
	"net/http"
	"time"

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

	// Try to restore metrics from file
	if config.restore {
		err := storage.RestoreMetricsFromFile(memStorage, config.fileStoragePath)
		if err != nil {
			zap.L().Error("Error restoring metrics", zap.Error(err))
		}
	}

	// Save metrics to file every {config.storeInterval} seconds
	storeToFileTicker := time.NewTicker(time.Duration(config.storeInterval) * time.Second)
	go func() {
		for {
			<-storeToFileTicker.C
			_ = storage.WriteMetricsToFile(memStorage, config.fileStoragePath)
		}
	}()

	err := http.ListenAndServe(config.serverAddress, r)
	if err != nil {
		zap.L().Fatal("Failed to start server", zap.Error(err))
	}
}
