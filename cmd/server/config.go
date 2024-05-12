package main

import (
	"flag"

	"github.com/caarlos0/env/v11"
	"github.com/zavtra-na-rabotu/gometrics/internal"
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
		internal.ErrorLog.Printf("Failed to parse environment variables: %s", err)
	}

	if cfg.Address != "" {
		serverAddress = cfg.Address
	}
}
