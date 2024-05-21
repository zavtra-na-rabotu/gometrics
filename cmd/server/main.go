package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zavtra-na-rabotu/gometrics/internal/logger"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/handlers"
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

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetric(memStorage))
	r.Get("/value/{type}/{name}", handlers.GetMetric(memStorage))
	r.Get("/", handlers.RenderAllMetrics(memStorage))

	err := http.ListenAndServe(serverAddress, r)
	if err != nil {
		zap.L().Fatal("Failed to start server", zap.Error(err))
	}
}
