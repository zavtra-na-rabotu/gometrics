package storage

import "errors"

type Storage interface {
	UpdateGauge(name string, metric float64) error
	UpdateCounter(name string, metric int64) error
	UpdateCounterAndReturn(name string, metric int64) (int64, error)
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauge() (map[string]float64, error)
	GetAllCounter() (map[string]int64, error)
}

var ErrItemNotFound = errors.New("item not found")
