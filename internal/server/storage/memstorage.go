package storage

import (
	"fmt"
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

func (storage *MemStorage) GetGauge(name string) (float64, error) {
	value, ok := storage.gauge[name]
	if !ok {
		return 0, fmt.Errorf("gauge metric with name: %s not found %w", name, ErrorItemNotFound)
	}

	return value, nil
}

func (storage *MemStorage) GetCounter(name string) (int64, error) {
	value, ok := storage.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter metric with name: %s not found %w", name, ErrorItemNotFound)
	}

	return value, nil
}

func (storage *MemStorage) GetAllGauge() map[string]float64 {
	return storage.gauge
}

func (storage *MemStorage) GetAllCounter() map[string]int64 {
	return storage.counter
}
