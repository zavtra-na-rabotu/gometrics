package storage

type Storage interface {
	UpdateGauge(name string, metric float64)
	UpdateCounter(name string, metric int64)
}
