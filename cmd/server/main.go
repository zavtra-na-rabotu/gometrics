package main

import (
	"log"
	"net/http"
	"strconv"
)

type MetricType string

const (
	gauge   MetricType = "gauge"
	counter MetricType = "counter"
)

type Storage interface {
	UpdateGauge(name string, metric string) error
	UpdateCounter(name string, metric string) error
}

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{make(map[string]float64), make(map[string]int64)}
}

func (storage *MemStorage) UpdateGauge(name string, metric string) error {
	value, err := strconv.ParseFloat(metric, 64)
	if err != nil {
		log.Printf("ERROR: failed to parse value from metric '%s': %s\n", metric, err)
		return err
	}
	storage.gauge[name] = value
	log.Printf("updated gauge: '%s' with value: %f", name, value)
	return nil
}

func (storage *MemStorage) UpdateCounter(name string, metric string) error {
	value, err := strconv.ParseInt(metric, 10, 64)
	if err != nil {
		log.Printf("ERROR: failed to parse value from metric '%s': %s\n", metric, err)
		return err
	}
	storage.counter[name] += value
	log.Printf("updated counter: '%s' with value: %d", name, value)
	return nil
}

func MetricHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate request method
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Validate Content-Type
		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Get data from request
		metricType := MetricType(r.PathValue("type"))
		metricName := r.PathValue("name")
		metricValue := r.PathValue("value")

		// Validate metricType TODO: (Move to validation pkg)
		if !(metricType == counter || metricType == gauge) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate metricName TODO: (Move to validation pkg)
		if len(metricName) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch metricType {
		case counter:
			err := storage.UpdateCounter(metricName, metricValue)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case gauge:
			err := storage.UpdateGauge(metricName, metricValue)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	storage := NewMemStorage()

	mux := http.NewServeMux()

	mux.HandleFunc("/update/{type}/{name}/{value}", MetricHandler(storage))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
