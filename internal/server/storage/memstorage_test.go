package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	storage := NewMemStorage()

	err := storage.UpdateGauge("TestGauge", 10.5)
	assert.NoError(t, err)

	value, err := storage.GetGauge("TestGauge")
	assert.NoError(t, err)
	assert.Equal(t, 10.5, value)
}

func TestMemStorage_GetGauge_NotFound(t *testing.T) {
	storage := NewMemStorage()

	_, err := storage.GetGauge("NonExistingGauge")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrItemNotFound))
}

func TestMemStorage_GetAllGauge(t *testing.T) {
	storage := NewMemStorage()

	expectedGauge1 := 10.1
	expectedGauge2 := 20.2

	err := storage.UpdateGauge("gauge1", expectedGauge1)
	assert.NoError(t, err)

	err = storage.UpdateGauge("gauge2", expectedGauge2)
	assert.NoError(t, err)

	gauges, err := storage.GetAllGauge()
	assert.NoError(t, err)

	assert.Equal(t, 2, len(gauges))
	assert.Equal(t, expectedGauge1, gauges["gauge1"])
	assert.Equal(t, expectedGauge2, gauges["gauge2"])
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	storage := NewMemStorage()

	err := storage.UpdateCounter("TestCounter", 5)
	assert.NoError(t, err)

	value, err := storage.GetCounter("TestCounter")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), value)

	err = storage.UpdateCounter("TestCounter", 3)
	assert.NoError(t, err)

	value, err = storage.GetCounter("TestCounter")
	assert.NoError(t, err)
	assert.Equal(t, int64(8), value)
}

func TestMemStorage_GetCounter_NotFound(t *testing.T) {
	storage := NewMemStorage()

	_, err := storage.GetCounter("NonExistingCounter")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrItemNotFound))
}

func TestMemStorage_UpdateCounterAndReturn(t *testing.T) {
	storage := NewMemStorage()

	value, err := storage.UpdateCounterAndReturn("TestCounter", 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), value)

	value, err = storage.UpdateCounterAndReturn("TestCounter", 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), value)
}

func TestMemStorage_GetAllCounter(t *testing.T) {
	storage := NewMemStorage()

	expectedCounter1 := int64(10)
	expectedCounter2 := int64(20)

	err := storage.UpdateCounter("counter1", expectedCounter1)
	assert.NoError(t, err)

	err = storage.UpdateCounter("counter2", expectedCounter2)
	assert.NoError(t, err)

	counters, err := storage.GetAllCounter()
	assert.NoError(t, err)

	assert.Equal(t, 2, len(counters))
	assert.Equal(t, expectedCounter1, counters["counter1"])
	assert.Equal(t, expectedCounter2, counters["counter2"])
}

func TestMemStorage_UpdateMetrics(t *testing.T) {
	storage := NewMemStorage()

	expectedGauge := 123.4
	expectedCounter := int64(12)

	metrics := []model.Metrics{
		{ID: "TestGauge", MType: string(model.Gauge), Value: &expectedGauge},
		{ID: "TestCounter", MType: string(model.Counter), Delta: &expectedCounter},
	}

	err := storage.UpdateMetrics(metrics)
	assert.NoError(t, err)

	gauge, err := storage.GetGauge("TestGauge")
	assert.NoError(t, err)
	assert.Equal(t, expectedGauge, gauge)

	counter, err := storage.GetCounter("TestCounter")
	assert.NoError(t, err)
	assert.Equal(t, expectedCounter, counter)
}
