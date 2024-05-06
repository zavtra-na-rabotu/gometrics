package main

import "flag"

var serverAddress string
var reportInterval int
var pollInterval int

func ParseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "host and port of server")
	flag.IntVar(&reportInterval, "r", 10, "report interval time in sec")
	flag.IntVar(&pollInterval, "p", 2, "poll interval time in sec")

	flag.Parse()
}
