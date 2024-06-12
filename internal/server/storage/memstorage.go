package storage

import (
	"fmt"
	"sync"

	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"go.uber.org/zap"
)

type MemStorage struct {
	gauge       map[string]float64
	counter     map[string]int64
	fileWriter  *Writer
	gaugeLock   sync.RWMutex
	counterLock sync.RWMutex
	syncMode    bool
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		make(map[string]float64),
		make(map[string]int64),
		nil,
		sync.RWMutex{},
		sync.RWMutex{},
		false,
	}
}

func (storage *MemStorage) SetSyncMode(syncMode bool) {
	storage.syncMode = syncMode
}

func (storage *MemStorage) SetFileWriter(fileWriter *Writer) {
	storage.fileWriter = fileWriter
}

func (storage *MemStorage) UpdateGauge(name string, metric float64) error {
	storage.gaugeLock.Lock()
	defer storage.gaugeLock.Unlock()

	storage.gauge[name] = metric
	if storage.syncMode {
		err := storage.fileWriter.WriteMetric(model.Metrics{ID: name, MType: string(model.Gauge), Value: &metric})
		if err != nil {
			zap.L().Error("Failed to write gauge to file", zap.String("name", name), zap.Float64("metric", metric), zap.Error(err))
			return fmt.Errorf("failed to write gauge to file: %w", err)
		}
	}
	zap.L().Info("Updated gauge", zap.String("name", name), zap.Float64("metric", metric))
	return nil
}

func (storage *MemStorage) UpdateCounter(name string, metric int64) error {
	storage.counterLock.Lock()
	defer storage.counterLock.Unlock()

	storage.counter[name] += metric
	metric = storage.counter[name]
	if storage.syncMode {
		err := storage.fileWriter.WriteMetric(model.Metrics{ID: name, MType: string(model.Counter), Delta: &metric})
		if err != nil {
			zap.L().Error("Failed to write counter to file", zap.String("name", name), zap.Int64("metric", metric), zap.Error(err))
			return fmt.Errorf("failed to write counter to file: %w", err)
		}
	}
	zap.L().Info("Updated counter", zap.String("name", name), zap.Int64("metric", metric))
	return nil
}

func (storage *MemStorage) UpdateCounterAndReturn(name string, metric int64) (int64, error) {
	storage.counterLock.Lock()
	defer storage.counterLock.Unlock()

	storage.counter[name] += metric
	metric = storage.counter[name]
	if storage.syncMode {
		err := storage.fileWriter.WriteMetric(model.Metrics{ID: name, MType: string(model.Counter), Delta: &metric})
		if err != nil {
			zap.L().Error("Failed to write counter to file", zap.String("name", name), zap.Int64("metric", metric), zap.Error(err))
			return metric, fmt.Errorf("failed to write counter to file: %w", err)
		}
	}
	zap.L().Info("Updated counter", zap.String("name", name), zap.Int64("metric", metric))

	return storage.counter[name], nil
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

func (storage *MemStorage) GetAllGauge() (map[string]float64, error) {
	storage.gaugeLock.RLock()
	defer storage.gaugeLock.RUnlock()

	gaugeCopy := make(map[string]float64, len(storage.gauge))
	for key, value := range storage.gauge {
		gaugeCopy[key] = value
	}

	return gaugeCopy, nil
}

func (storage *MemStorage) GetAllCounter() (map[string]int64, error) {
	storage.counterLock.RLock()
	defer storage.counterLock.RUnlock()

	counterCopy := make(map[string]int64, len(storage.counter))
	for key, value := range storage.counter {
		counterCopy[key] = value
	}

	return counterCopy, nil
}

func (storage *MemStorage) UpdateMetrics(metrics []model.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case string(model.Gauge):
			err := storage.UpdateGauge(metric.ID, *metric.Value)
			if err != nil {
				return err
			}
		case string(model.Counter):
			err := storage.UpdateCounter(metric.ID, *metric.Delta)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown metric type: %s", metric.MType)
		}
	}

	return nil
}
