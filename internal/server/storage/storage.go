package storage

import "errors"

type Storage interface {
	UpdateGauge(name string, metric float64)
	UpdateCounter(name string, metric int64)
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauge() map[string]float64
	GetAllCounter() map[string]int64
}

var ErrorItemNotFound = errors.New("item not found")
