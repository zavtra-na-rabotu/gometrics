package metrics

import (
	"math/rand/v2"
	"runtime"
)

type Collector struct {
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
}

func NewCollector() *Collector {
	return &Collector{
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
	}
}

func (metrics *Collector) ResetPollCounter() {
	metrics.counterMetrics["PollCount"] = 0
}

func (metrics *Collector) Collect() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics.gaugeMetrics["Alloc"] = float64(memStats.Alloc)
	metrics.gaugeMetrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	metrics.gaugeMetrics["Frees"] = float64(memStats.Frees)
	metrics.gaugeMetrics["GCCPUFraction"] = memStats.GCCPUFraction
	metrics.gaugeMetrics["GCSys"] = float64(memStats.GCSys)
	metrics.gaugeMetrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	metrics.gaugeMetrics["HeapIdle"] = float64(memStats.HeapIdle)
	metrics.gaugeMetrics["HeapInuse"] = float64(memStats.HeapInuse)
	metrics.gaugeMetrics["HeapObjects"] = float64(memStats.HeapObjects)
	metrics.gaugeMetrics["HeapReleased"] = float64(memStats.HeapReleased)
	metrics.gaugeMetrics["HeapSys"] = float64(memStats.HeapSys)
	metrics.gaugeMetrics["LastGC"] = float64(memStats.LastGC)
	metrics.gaugeMetrics["Lookups"] = float64(memStats.Lookups)
	metrics.gaugeMetrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	metrics.gaugeMetrics["MCacheSys"] = float64(memStats.MCacheSys)
	metrics.gaugeMetrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	metrics.gaugeMetrics["MSpanSys"] = float64(memStats.MSpanSys)
	metrics.gaugeMetrics["Mallocs"] = float64(memStats.Mallocs)
	metrics.gaugeMetrics["NextGC"] = float64(memStats.NextGC)
	metrics.gaugeMetrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	metrics.gaugeMetrics["NumGC"] = float64(memStats.NumGC)
	metrics.gaugeMetrics["OtherSys"] = float64(memStats.OtherSys)
	metrics.gaugeMetrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	metrics.gaugeMetrics["StackInuse"] = float64(memStats.StackInuse)
	metrics.gaugeMetrics["StackSys"] = float64(memStats.StackSys)
	metrics.gaugeMetrics["Sys"] = float64(memStats.Sys)
	metrics.gaugeMetrics["TotalAlloc"] = float64(memStats.TotalAlloc)
	metrics.gaugeMetrics["RandomValue"] = rand.Float64()
	metrics.counterMetrics["PollCount"]++
}
