package metrics

import (
	"sync"
	"time"

	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"go.uber.org/zap"
)

// SenderClient client interface (HTTP and gRPC implemented)
type SenderClient interface {
	SendMetrics(metrics []model.Metrics) error
}

// Sender structure with all dependencies for metrics sending
type Sender struct {
	collector      *Collector
	rateLimit      int
	reportInterval time.Duration
	client         SenderClient
}

// NewSender sender constructor
func NewSender(rateLimit int, reportInterval int, collector *Collector, client SenderClient) *Sender {
	return &Sender{
		collector:      collector,
		rateLimit:      rateLimit,
		reportInterval: time.Duration(reportInterval) * time.Second,
		client:         client,
	}
}

// InitSender init and start sending collected metrics with specified interval and rate limit
func (sender *Sender) InitSender() {
	sendJobs := make(chan []model.Metrics, sender.rateLimit)

	var wg sync.WaitGroup
	for i := 1; i <= sender.rateLimit; i++ {
		wg.Add(1)
		go sender.worker(i, sendJobs, &wg)
	}

	ticker := time.NewTicker(sender.reportInterval)
	defer ticker.Stop()

	for range ticker.C {
		zap.L().Info("Sending metrics")
		metric, ok := <-sender.collector.metrics
		if !ok {
			close(sendJobs)
			wg.Wait()
			return
		}
		sendJobs <- metric
	}
}

func (sender *Sender) worker(id int, jobs <-chan []model.Metrics, wg *sync.WaitGroup) {
	zap.L().Info("Starting worker", zap.Int("Worker id", id))
	defer wg.Done()

	for metrics := range jobs {
		err := sender.client.SendMetrics(metrics)
		if err != nil {
			zap.L().Error("Error sending metrics", zap.Error(err), zap.Int("Worker id", id))
		} else {
			sender.collector.ResetPollCounter()
		}
	}
}
