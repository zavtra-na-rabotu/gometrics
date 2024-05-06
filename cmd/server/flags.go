package main

import "flag"

var serverAddress string

func ParseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "host and port of server")

	flag.Parse()
}
