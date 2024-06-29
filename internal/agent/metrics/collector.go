package metrics

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"go.uber.org/zap"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
)

type Collector struct {
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
	pollInterval   time.Duration
	metrics        chan []model.Metrics
	gaugeLock      sync.RWMutex
	counterLock    sync.RWMutex
}

func NewCollector(pollInterval int, metrics chan []model.Metrics) *Collector {
	return &Collector{
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
		pollInterval:   time.Duration(pollInterval) * time.Second,
		metrics:        metrics,
		gaugeLock:      sync.RWMutex{},
		counterLock:    sync.RWMutex{},
	}
}

func (collector *Collector) ResetPollCounter() {
	collector.counterLock.Lock()
	defer collector.counterLock.Unlock()

	collector.counterMetrics["PollCount"] = 0
}

func (collector *Collector) InitCollector() {
	ticker := time.NewTicker(collector.pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		collector.CollectGaugeMetrics()
		collector.CollectCounterMetrics()

		metrics := collector.CreateMetrics()

		select {
		case collector.metrics <- metrics:
		default:
			select {
			case <-collector.metrics:
				zap.L().Info("Cleared old metrics from channel")
			default:
				// Channel is empty
			}

			select {
			case collector.metrics <- metrics:
				zap.L().Info("Sent new metrics after clearing")
			default:
				// Channel still blocked
			}
		}
	}
}

func (collector *Collector) InitPsutilCollector() {
	ticker := time.NewTicker(collector.pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		collector.CollectPsutilMetrics()
	}
}

func (collector *Collector) CollectPsutilMetrics() {
	collector.gaugeLock.Lock()
	defer collector.gaugeLock.Unlock()

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		zap.L().Error("Failed to get memory info", zap.Error(err))
		return
	}

	collector.gaugeMetrics["TotalMemory"] = float64(memInfo.Total)
	collector.gaugeMetrics["FreeMemory"] = float64(memInfo.Free)

	cpuPercents, err := cpu.Percent(time.Second, true)
	if err != nil {
		zap.L().Error("Failed to get cpu percents", zap.Error(err))
		return
	}

	for cpuNum, cpuPercent := range cpuPercents {
		collector.gaugeMetrics["CPUutilization"+strconv.FormatInt(int64(cpuNum), 10)] = cpuPercent
	}
}

func (collector *Collector) CollectCounterMetrics() {
	collector.counterLock.Lock()
	defer collector.counterLock.Unlock()

	collector.counterMetrics["PollCount"]++
}

func (collector *Collector) CollectGaugeMetrics() {
	collector.gaugeLock.Lock()
	defer collector.gaugeLock.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	collector.gaugeMetrics["Alloc"] = float64(memStats.Alloc)
	collector.gaugeMetrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	collector.gaugeMetrics["Frees"] = float64(memStats.Frees)
	collector.gaugeMetrics["GCCPUFraction"] = memStats.GCCPUFraction
	collector.gaugeMetrics["GCSys"] = float64(memStats.GCSys)
	collector.gaugeMetrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	collector.gaugeMetrics["HeapIdle"] = float64(memStats.HeapIdle)
	collector.gaugeMetrics["HeapInuse"] = float64(memStats.HeapInuse)
	collector.gaugeMetrics["HeapObjects"] = float64(memStats.HeapObjects)
	collector.gaugeMetrics["HeapReleased"] = float64(memStats.HeapReleased)
	collector.gaugeMetrics["HeapSys"] = float64(memStats.HeapSys)
	collector.gaugeMetrics["LastGC"] = float64(memStats.LastGC)
	collector.gaugeMetrics["Lookups"] = float64(memStats.Lookups)
	collector.gaugeMetrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	collector.gaugeMetrics["MCacheSys"] = float64(memStats.MCacheSys)
	collector.gaugeMetrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	collector.gaugeMetrics["MSpanSys"] = float64(memStats.MSpanSys)
	collector.gaugeMetrics["Mallocs"] = float64(memStats.Mallocs)
	collector.gaugeMetrics["NextGC"] = float64(memStats.NextGC)
	collector.gaugeMetrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	collector.gaugeMetrics["NumGC"] = float64(memStats.NumGC)
	collector.gaugeMetrics["OtherSys"] = float64(memStats.OtherSys)
	collector.gaugeMetrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	collector.gaugeMetrics["StackInuse"] = float64(memStats.StackInuse)
	collector.gaugeMetrics["StackSys"] = float64(memStats.StackSys)
	collector.gaugeMetrics["Sys"] = float64(memStats.Sys)
	collector.gaugeMetrics["TotalAlloc"] = float64(memStats.TotalAlloc)
	collector.gaugeMetrics["RandomValue"] = rand.Float64()
}

func (collector *Collector) CreateMetrics() []model.Metrics {
	zap.L().Info("PollCount", zap.Int64("PollCount", collector.counterMetrics["PollCount"]))

	var metrics []model.Metrics

	collector.gaugeLock.RLock()
	for name, value := range collector.gaugeMetrics {
		metrics = append(metrics, *createMetric(name, string(model.Gauge), value, 0))
	}
	collector.gaugeLock.RUnlock()

	collector.counterLock.RLock()
	for name, value := range collector.counterMetrics {
		metrics = append(metrics, *createMetric(name, string(model.Counter), 0, value))
	}
	collector.counterLock.RUnlock()

	return metrics
}

func createMetric(id string, mType string, value float64, delta int64) *model.Metrics {
	return &model.Metrics{
		ID:    id,
		MType: mType,
		Value: &value,
		Delta: &delta,
	}
}
