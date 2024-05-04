package main

import (
	"github.com/zavtra-na-rabotu/gometrics/internal/server/handlers"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"log"
	"net/http"
)

func main() {
	memStorage := storage.NewMemStorage()

	mux := http.NewServeMux()

	mux.HandleFunc("/update/{type}/{name}/{value}", handlers.UpdateMetricHandler(memStorage))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
