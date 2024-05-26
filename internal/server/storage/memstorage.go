package storage

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type MemStorage struct {
	gauge       map[string]float64
	counter     map[string]int64
	gaugeLock   sync.RWMutex
	counterLock sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		make(map[string]float64),
		make(map[string]int64),
		sync.RWMutex{},
		sync.RWMutex{},
	}
}

func (storage *MemStorage) UpdateGauge(name string, metric float64) {
	storage.gaugeLock.Lock()
	defer storage.gaugeLock.Unlock()

	storage.gauge[name] = metric
	zap.L().Info("Updated gauge", zap.String("name", name), zap.Float64("metric", metric))
}

func (storage *MemStorage) UpdateCounter(name string, metric int64) {
	storage.counterLock.Lock()
	defer storage.counterLock.Unlock()

	storage.counter[name] += metric
	zap.L().Info("Updated counter", zap.String("name", name), zap.Int64("metric", metric))
}

func (storage *MemStorage) UpdateCounterAndReturn(name string, metric int64) int64 {
	storage.counterLock.Lock()
	defer storage.counterLock.Unlock()

	storage.counter[name] += metric
	zap.L().Info("Updated counter", zap.String("name", name), zap.Int64("metric", metric))

	return storage.counter[name]
}

func (storage *MemStorage) GetGauge(name string) (float64, error) {
	storage.gaugeLock.RLock()
	defer storage.gaugeLock.RUnlock()

	value, ok := storage.gauge[name]
	if !ok {
		return 0, fmt.Errorf("gauge metric with name: %s not found %w", name, ErrItemNotFound)
	}

	return value, nil
}

func (storage *MemStorage) GetCounter(name string) (int64, error) {
	storage.counterLock.RLock()
	defer storage.counterLock.RUnlock()

	value, ok := storage.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter metric with name: %s not found %w", name, ErrItemNotFound)
	}

	return value, nil
}

func (storage *MemStorage) GetAllGauge() map[string]float64 {
	storage.gaugeLock.RLock()
	defer storage.gaugeLock.RUnlock()

	gaugeCopy := make(map[string]float64, len(storage.gauge))
	for key, value := range storage.gauge {
		gaugeCopy[key] = value
	}

	return gaugeCopy
}

func (storage *MemStorage) GetAllCounter() map[string]int64 {
	storage.counterLock.RLock()
	defer storage.counterLock.RUnlock()

	counterCopy := make(map[string]int64, len(storage.counter))
	for key, value := range storage.counter {
		counterCopy[key] = value
	}

	return counterCopy
}
