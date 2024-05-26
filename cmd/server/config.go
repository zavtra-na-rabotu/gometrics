package main

import (
	"flag"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

var serverAddress string

type envs struct {
	Address string `env:"ADDRESS"`
}

func Configure() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "Server URL")
	flag.Parse()

	cfg := envs{}
	err := env.Parse(&cfg)
	if err != nil {
		zap.L().Error("Failed to parse environment variables", zap.Error(err))
	}

	if cfg.Address != "" {
		serverAddress = cfg.Address
	}
}
