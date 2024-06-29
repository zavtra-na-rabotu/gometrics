package metrics

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"go.uber.org/zap"
)

type Sender struct {
	client         *resty.Client
	collector      *Collector
	key            string
	rateLimit      int
	reportInterval time.Duration
}

func NewSender(url string, key string, rateLimit int, reportInterval int, collector *Collector) *Sender {
	return &Sender{
		client:         resty.New().SetBaseURL("http://" + url),
		collector:      collector,
		key:            key,
		rateLimit:      rateLimit,
		reportInterval: time.Duration(reportInterval) * time.Second,
	}
}

func (sender *Sender) worker(id int, jobs <-chan []model.Metrics, wg *sync.WaitGroup) {
	zap.L().Info("Starting worker", zap.Int("Worker id", id))
	defer wg.Done()

	for metrics := range jobs {
		//zap.L().Info("Sending metrics", zap.Any("metrics", metrics))
		err := sender.sendMetrics(metrics)
		if err != nil {
			zap.L().Error("Error sending metrics", zap.Error(err), zap.Int("Worker id", id))
		} else {
			sender.collector.ResetPollCounter()
		}
	}
}

func (sender *Sender) InitSender() {
	sendJobs := make(chan []model.Metrics, sender.rateLimit)
	defer close(sendJobs)

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
			wg.Wait()
			return
		}
		sendJobs <- metric
	}
}

func (sender *Sender) sendMetrics(metrics []model.Metrics) error {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics:  %w", err)
	}

	var compressedBody bytes.Buffer
	err = compressBody(&compressedBody, jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}

	request := sender.client.R()

	if sender.key != "" {
		var hash = calculateHash(jsonData, sender.key)
		request.SetHeader("HashSHA256", hash)
	}

	response, err := request.
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(compressedBody.Bytes()).
		Post("/updates/")

	if err != nil {
		return fmt.Errorf("failed to send metric %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send metric, StatusCode: %d", response.StatusCode())
	}
	return nil
}

func compressBody(compressedData *bytes.Buffer, jsonData []byte) error {
	gzipWriter := gzip.NewWriter(compressedData)
	_, err := gzipWriter.Write(jsonData)
	if err != nil {
		zap.L().Error("Error compressing data", zap.Error(err))
		return fmt.Errorf("error compressing data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		zap.L().Error("Error closing gzip writer", zap.Error(err))
		return fmt.Errorf("error closing gzip writer: %w", err)
	}

	return nil
}

func calculateHash(jsonData []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(jsonData)
	return hex.EncodeToString(hash.Sum(nil))
}
