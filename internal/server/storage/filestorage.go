package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"go.uber.org/zap"
)

type Writer struct {
	file   *os.File
	writer *bufio.Writer
}

var ErrFileStoragePathNotProvided = errors.New("file storage path not provided")

func NewWriter(fileStoragePath string) (*Writer, error) {
	file, err := os.OpenFile(fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	return &Writer{file, bufio.NewWriter(file)}, nil
}

func (p *Writer) WriteMetric(metrics model.Metrics) error {
	data, err := json.Marshal(&metrics)
	if err != nil {
		return fmt.Errorf("error marshalling metrics: %w", err)
	}
	if _, err := p.writer.Write(data); err != nil {
		return fmt.Errorf("error writing metrics: %w", err)
	}
	if err := p.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("error writing new line: %w", err)
	}
	return p.writer.Flush()
}

func (p *Writer) Close() error {
	return p.file.Close()
}

type Reader struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewReader(filename string) (*Reader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	return &Reader{file, bufio.NewScanner(file)}, nil
}

func (c *Reader) ReadMetric() (*model.Metrics, error) {
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}

	data := c.scanner.Bytes()

	metrics := model.Metrics{}
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling metrics: %w", err)
	}

	return &metrics, nil
}

func (c *Reader) Close() error {
	return c.file.Close()
}

func WriteMetricsToFile(memStorage *MemStorage, fileStoragePath string) error {
	if fileStoragePath == "" {
		return ErrFileStoragePathNotProvided
	}

	writer, err := NewWriter(fileStoragePath)
	if err != nil {
		return fmt.Errorf("could not create file writer: %w", err)
	}

	gaugeMetrics := memStorage.GetAllGauge()
	counterMetrics := memStorage.GetAllCounter()

	for name, metric := range gaugeMetrics {
		err = writer.WriteMetric(model.Metrics{ID: name, MType: string(model.Gauge), Value: &metric})
		if err != nil {
			return fmt.Errorf("could not write gauge metric: %w", err)
		}
	}

	for name, metric := range counterMetrics {
		err = writer.WriteMetric(model.Metrics{ID: name, MType: string(model.Counter), Delta: &metric})
		if err != nil {
			return fmt.Errorf("could not write counter metric: %w", err)
		}
	}

	return nil
}

func RestoreMetricsFromFile(memStorage *MemStorage, fileStoragePath string) error {
	if fileStoragePath == "" {
		return ErrFileStoragePathNotProvided
	}

	reader, err := NewReader(fileStoragePath)
	if err != nil {
		zap.L().Error("Failed to create new reader", zap.Error(err))
		return fmt.Errorf("failed to create new reader: %w", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			zap.L().Error("failed to close reader", zap.Error(err))
		}
	}()

	for {
		metric, err := reader.ReadMetric()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			continue
		}
		if metric == nil {
			break
		}
		if metric.MType == string(model.Gauge) {
			err = memStorage.UpdateGauge(metric.ID, *metric.Value)
			if err != nil {
				return fmt.Errorf("could not update gauge: %w", err)
			}
			zap.L().Info("Gauge read from file", zap.String("name", metric.ID), zap.Float64("value", *metric.Value))
		} else if metric.MType == string(model.Counter) {
			err = memStorage.UpdateCounter(metric.ID, *metric.Delta)
			if err != nil {
				return fmt.Errorf("could not update counter: %w", err)
			}
			zap.L().Info("Counter read from file", zap.String("name", metric.ID), zap.Int64("delta", *metric.Delta))
		}
	}

	return nil
}
