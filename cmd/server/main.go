package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	server := &http.Server{
		Addr:    config.ServerAddress,
		Handler: r,
	}

	// Channel for signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Start server in separate goroutine
	go func() {
		zap.L().Info("Starting server", zap.String("address", config.ServerAddress))

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Waiting for signal
	sig := <-quit
	zap.L().Info("Shutting down server...", zap.String("signal", sig.String()))

	// Context with timeout to shut down server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		zap.L().Fatal("Server forced to shutdown", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}
