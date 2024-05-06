package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/handlers"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"log"
	"net/http"
)

func main() {
	memStorage := storage.NewMemStorage()
	r := chi.NewRouter()

	// TODO: Split to different handlers based on type (/update/gauge/*, /update/counter/*)
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetric(memStorage))
	r.Get("/value/{type}/{name}", handlers.GetMetric(memStorage))
	r.Get("/", handlers.RenderAllMetrics(memStorage))

	ParseFlags()

	log.Fatal(http.ListenAndServe(serverAddress, r))
}
