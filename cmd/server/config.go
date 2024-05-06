package main

import (
	"flag"
	"os"
)

var serverAddress string

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "host and port of server")

	flag.Parse()
}

func readEnvVariables() {
	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		serverAddress = envServerAddress
	}
}

func Configure() {
	parseFlags()
	readEnvVariables()
}
