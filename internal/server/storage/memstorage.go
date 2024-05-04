package storage

import (
	"log"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		make(map[string]float64),
		make(map[string]int64),
	}
}

func (storage *MemStorage) UpdateGauge(name string, metric float64) {
	storage.gauge[name] = metric
	log.Printf("updated gauge: '%s' with value: %f", name, metric)
}

func (storage *MemStorage) UpdateCounter(name string, metric int64) {
	storage.counter[name] += metric
	log.Printf("updated counter: '%s' with value: %d", name, metric)
}
