package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
)

func TestNewWriter(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	writer, err := NewWriter(tempFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, writer)

	err = writer.Close()
	assert.NoError(t, err)
}

func TestNewReader(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	reader, err := NewReader(tempFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	err = reader.Close()
	assert.NoError(t, err)
}

func TestWriter_WriteMetric_ReadMetric(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	writer, err := NewWriter(tempFile.Name())
	assert.NoError(t, err)
	defer func() {
		err = writer.Close()
		assert.NoError(t, err)
	}()

	reader, err := NewReader(tempFile.Name())
	assert.NoError(t, err)
	defer func() {
		err = reader.Close()
		assert.NoError(t, err)
	}()

	value := 123.45
	metric := model.Metrics{
		ID:    "TestGauge",
		MType: string(model.Gauge),
		Value: &value,
	}

	err = writer.WriteMetric(metric)
	assert.NoError(t, err)

	readMetric, err := reader.ReadMetric()
	assert.NoError(t, err)
	assert.Equal(t, metric.ID, readMetric.ID)
	assert.Equal(t, metric.MType, readMetric.MType)
	assert.Equal(t, *metric.Value, *readMetric.Value)
}

func TestWriteMetricsToFile_RestoreMetricsFromFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	expectedGauge := 123.45
	expectedCounter := int64(10)

	writerMemStorage := NewMemStorage()
	writerMemStorage.UpdateGauge("gauge_metric", expectedGauge)
	writerMemStorage.UpdateCounter("counter_metric", expectedCounter)

	err = WriteMetricsToFile(writerMemStorage, tempFile.Name())
	assert.NoError(t, err)

	readerMemStorage := NewMemStorage()
	err = RestoreMetricsFromFile(readerMemStorage, tempFile.Name())
	assert.NoError(t, err)

	gauge, _ := readerMemStorage.GetGauge("gauge_metric")
	assert.Equal(t, expectedGauge, gauge)

	counter, _ := readerMemStorage.GetCounter("counter_metric")
	assert.Equal(t, expectedCounter, counter)
}
