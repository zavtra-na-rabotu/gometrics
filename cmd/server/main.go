package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	profilermiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/zavtra-na-rabotu/gometrics/internal/logger"
	"github.com/zavtra-na-rabotu/gometrics/internal/pb"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/configuration"
	grpcv1 "github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/grpc/v1"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/http/v1"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/http/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/handlers/http/v3"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/interceptor"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/middleware"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/security"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	logger.InitLogger()

	sugar := logger.GetSugaredLogger()
	sugar.Infof("Build version: %s", buildVersion)
	sugar.Infof("Build date: %s", buildDate)
	sugar.Infof("Build commit: %s", buildCommit)

	config := configuration.Configure()

	// Configure storage
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

	// Configure trusted ipnet
	var trustedIPNet *net.IPNet
	if config.TrustedSubnet != "" {
		_, ipnet, err := net.ParseCIDR(config.TrustedSubnet)
		if err != nil {
			zap.L().Fatal("Failed to parse trusted subnet", zap.Error(err))
		}

		trustedIPNet = ipnet
	}

	if config.GRPCEnabled && config.GRPCPort >= 1024 && config.GRPCPort <= 65535 {
		startGRPC(config, storageToUse, trustedIPNet)
	} else {
		zap.L().Info("gRPC disabled or port is wrong, fallback to HTTP", zap.Bool("grpc_enabled", config.GRPCEnabled), zap.Int("grpc_port", config.GRPCPort))
		startHTTP(config, storageToUse, trustedIPNet)
	}
}

func startHTTP(config *configuration.Configuration, storage storage.Storage, ipnet *net.IPNet) {
	r := chi.NewRouter()

	if ipnet != nil {
		r.Use(middleware.IPValidation(ipnet))
	}

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

	// API v1
	r.Post("/update/{type}/{name}/{value}", v1.UpdateMetric(storage))
	r.Get("/value/{type}/{name}", v1.GetMetric(storage))
	r.Get("/", v1.RenderAllMetrics(storage))

	// API v2
	r.Post("/update/", v2.UpdateMetric(storage))
	r.Post("/value/", v2.GetMetric(storage))

	// API v3
	r.Get("/ping", v3.Ping(storage))
	r.Post("/updates/", v3.UpdateMetrics(storage))

	// Profiler
	r.Mount("/debug", profilermiddleware.Profiler())

	server := &http.Server{
		Addr:    config.ServerAddress,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// Start server in separate goroutine
	go func() {
		zap.L().Info("Starting HTTP server", zap.String("address", config.ServerAddress))

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("HTTP Server failed to start", zap.Error(err))
		}
	}()

	// Waiting for signal
	<-ctx.Done()
	zap.L().Info("Shutting down HTTP server...")

	// Context with timeout to shut down server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		zap.L().Fatal("HTTP Server forced to shutdown", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}

func startGRPC(config *configuration.Configuration, storage storage.Storage, ipnet *net.IPNet) {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GRPCPort))
	if err != nil {
		zap.L().Fatal("Failed to listen", zap.Error(err))
	}

	// Add needed server interceptors
	var interceptors []grpc.UnaryServerInterceptor

	if ipnet != nil {
		interceptors = append(interceptors, interceptor.IPValidationInterceptor(ipnet))
	}

	if config.Key != "" {
		interceptors = append(interceptors, interceptor.HashInterceptor(config.Key))
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	pb.RegisterMetricsServiceServer(gRPCServer, grpcv1.NewServer(storage))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// Start gRPC server in separate goroutine
	go func() {
		zap.L().Info("gRPC server started", zap.Int("port", config.GRPCPort))
		if err := gRPCServer.Serve(listen); err != nil {
			zap.L().Fatal("Failed to serve", zap.Error(err))
		}
	}()

	// Waiting for signal
	<-ctx.Done()
	zap.L().Info("Shutting down gRPC server...")
	gRPCServer.GracefulStop()
}
