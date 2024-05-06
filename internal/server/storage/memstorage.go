package storage

import (
	"fmt"
	"log"
	"sync"
)

// TODO: Вопрос для ревью, нужен ли лок на чтение, ибо используем большие типы int64 float64, чтение/запись не атомарна
type MemStorage struct {
	gaugeLock   sync.RWMutex
	counterLock sync.RWMutex
	gauge       map[string]float64
	counter     map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		sync.RWMutex{},
		sync.RWMutex{},
		make(map[string]float64),
		make(map[string]int64),
	}
}

func (storage *MemStorage) UpdateGauge(name string, metric float64) {
	storage.gaugeLock.Lock()
	defer storage.gaugeLock.Unlock()

	storage.gauge[name] = metric
	log.Printf("updated gauge: '%s' with value: %f", name, metric)
}

func (storage *MemStorage) UpdateCounter(name string, metric int64) {
	storage.counterLock.Lock()
	defer storage.counterLock.Unlock()

	storage.counter[name] += metric
	log.Printf("updated counter: '%s' with value: %d", name, metric)
}

func (storage *MemStorage) GetGauge(name string) (float64, error) {
	storage.gaugeLock.RLock()
	defer storage.gaugeLock.RUnlock()

	value, ok := storage.gauge[name]
	if !ok {
		return 0, fmt.Errorf("gauge metric with name: %s not found %w", name, ErrorItemNotFound)
	}

	return value, nil
}

func (storage *MemStorage) GetCounter(name string) (int64, error) {
	storage.counterLock.RLock()
	defer storage.counterLock.RUnlock()

	value, ok := storage.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter metric with name: %s not found %w", name, ErrorItemNotFound)
	}

	return value, nil
}

func (storage *MemStorage) GetAllGauge() map[string]float64 {
	storage.gaugeLock.RLock()
	defer storage.gaugeLock.RUnlock()

	return storage.gauge
}

func (storage *MemStorage) GetAllCounter() map[string]int64 {
	storage.counterLock.RLock()
	defer storage.counterLock.RUnlock()

	return storage.counter
}
