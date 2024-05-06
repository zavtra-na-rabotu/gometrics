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

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetric(memStorage))
	r.Get("/value/{type}/{name}", handlers.GetMetric(memStorage))
	r.Get("/", handlers.RenderAllMetrics(memStorage))

	log.Fatal(http.ListenAndServe(":8080", r))
}
