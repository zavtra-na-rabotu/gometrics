package main

import (
	"flag"
	"github.com/zavtra-na-rabotu/gometrics/internal"
	"os"
	"strconv"
)

var serverAddress string
var reportInterval int
var pollInterval int

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "host and port of server")
	flag.IntVar(&reportInterval, "r", 10, "report interval time in sec")
	flag.IntVar(&pollInterval, "p", 2, "poll interval time in sec")

	flag.Parse()
}

func readEnvVariables() {
	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		serverAddress = envServerAddress
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		parsedReportInterval, err := strconv.Atoi(envReportInterval)
		if err != nil {
			internal.ErrorLog.Printf("failed to parse REPORT_INTERVAL: %s", err)
		} else {
			reportInterval = parsedReportInterval
		}
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		parsedPollInterval, err := strconv.Atoi(envPollInterval)
		if err != nil {
			internal.ErrorLog.Printf("failed to parse POLL_INTERVAL: %s", err)
		} else {
			pollInterval = parsedPollInterval
		}
	}
}

func Configure() {
	parseFlags()
	readEnvVariables()
}
