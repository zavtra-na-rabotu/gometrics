package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
)

func TestCollector_NewCollector(t *testing.T) {
	metricsCh := make(chan []model.Metrics, 1)
	collector := NewCollector(1, metricsCh)

	assert.NotNil(t, collector)
	assert.Equal(t, time.Second, collector.pollInterval)
}

func TestCollector_ResetPollCounter(t *testing.T) {
	metricsCh := make(chan []model.Metrics, 1)
	collector := NewCollector(1, metricsCh)
	collector.counterMetrics["PollCount"] = 10

	collector.ResetPollCounter()

	assert.Equal(t, int64(0), collector.counterMetrics["PollCount"])
}

func TestCollector_CollectGaugeMetrics(t *testing.T) {
	metricsCh := make(chan []model.Metrics, 1)
	collector := NewCollector(1, metricsCh)

	collector.CollectGaugeMetrics()

	assert.NotZero(t, collector.gaugeMetrics["Alloc"])
	assert.NotZero(t, collector.gaugeMetrics["TotalAlloc"])
	assert.NotZero(t, collector.gaugeMetrics["RandomValue"])
	assert.NotZero(t, collector.gaugeMetrics["Alloc"])
	assert.NotZero(t, collector.gaugeMetrics["BuckHashSys"])
	assert.NotZero(t, collector.gaugeMetrics["Frees"])
	assert.NotZero(t, collector.gaugeMetrics["GCSys"])
	assert.NotZero(t, collector.gaugeMetrics["HeapAlloc"])
	assert.NotZero(t, collector.gaugeMetrics["HeapIdle"])
	assert.NotZero(t, collector.gaugeMetrics["HeapInuse"])
	assert.NotZero(t, collector.gaugeMetrics["HeapObjects"])
	assert.NotZero(t, collector.gaugeMetrics["HeapReleased"])
	assert.NotZero(t, collector.gaugeMetrics["HeapSys"])
	assert.NotZero(t, collector.gaugeMetrics["MCacheInuse"])
	assert.NotZero(t, collector.gaugeMetrics["MCacheSys"])
	assert.NotZero(t, collector.gaugeMetrics["MSpanInuse"])
	assert.NotZero(t, collector.gaugeMetrics["MSpanSys"])
	assert.NotZero(t, collector.gaugeMetrics["Mallocs"])
	assert.NotZero(t, collector.gaugeMetrics["NextGC"])
	assert.NotZero(t, collector.gaugeMetrics["OtherSys"])
	assert.NotZero(t, collector.gaugeMetrics["StackInuse"])
	assert.NotZero(t, collector.gaugeMetrics["StackSys"])
	assert.NotZero(t, collector.gaugeMetrics["Sys"])
	assert.NotZero(t, collector.gaugeMetrics["TotalAlloc"])
}

func TestCollector_CollectCounterMetrics(t *testing.T) {
	metricsCh := make(chan []model.Metrics, 1)
	collector := NewCollector(1, metricsCh)
	collector.counterMetrics["PollCount"] = 5

	collector.CollectCounterMetrics()

	assert.Equal(t, int64(6), collector.counterMetrics["PollCount"])
}

func TestCollector_CollectPsutilMetrics(t *testing.T) {
	metricsCh := make(chan []model.Metrics, 1)
	collector := NewCollector(1, metricsCh)

	collector.CollectPsutilMetrics()

	assert.Contains(t, collector.gaugeMetrics, "TotalMemory")
	assert.Contains(t, collector.gaugeMetrics, "FreeMemory")
}

func TestCollector_CreateMetrics(t *testing.T) {
	tests := []struct {
		name           string
		collector      *Collector
		gaugeMetrics   map[string]float64
		counterMetrics map[string]int64
	}{
		{
			"Create gauge and counter metric",
			NewCollector(1, make(chan []model.Metrics, 2)),
			map[string]float64{"TestGauge": 123.456},
			map[string]int64{"TestCounter": 123},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for name, value := range test.gaugeMetrics {
				test.collector.gaugeMetrics[name] = value
			}

			for name, value := range test.counterMetrics {
				test.collector.counterMetrics[name] = value
			}

			result := test.collector.CreateMetrics()

			assert.Len(t, result, len(test.collector.gaugeMetrics)+len(test.collector.counterMetrics))
		})
	}
}

func TestCollector_InitCollector(t *testing.T) {
	metricsCh := make(chan []model.Metrics, 1)
	collector := NewCollector(1, metricsCh)

	go collector.InitCollector()

	select {
	case metrics := <-metricsCh:
		assert.NotEmpty(t, metrics)
	case <-time.After(3 * time.Second):
		t.Error("Timeout: no metrics were sent to the channel")
	}
}

func TestCollector_InitPsutilCollector(t *testing.T) {
	metricsCh := make(chan []model.Metrics, 1)
	collector := NewCollector(1, metricsCh)

	go collector.InitPsutilCollector()

	time.Sleep(3 * time.Second)

	assert.NotZero(t, collector.gaugeMetrics["TotalMemory"])
	assert.NotZero(t, collector.gaugeMetrics["FreeMemory"])
}
