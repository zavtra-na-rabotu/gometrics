package metrics

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"go.uber.org/zap"
)

// Sender structure with all dependencies for metrics sending
type Sender struct {
	client         *resty.Client
	collector      *Collector
	key            string
	rateLimit      int
	reportInterval time.Duration
	publicKey      *rsa.PublicKey
	localIP        net.IP
}

// NewSender sender constructor
func NewSender(url string, key string, rateLimit int, reportInterval int, publicKey *rsa.PublicKey, collector *Collector) *Sender {
	return &Sender{
		client:         resty.New().SetBaseURL("http://" + url),
		collector:      collector,
		key:            key,
		rateLimit:      rateLimit,
		reportInterval: time.Duration(reportInterval) * time.Second,
		publicKey:      publicKey,
		localIP:        getLocalIP(),
	}
}

func (sender *Sender) worker(id int, jobs <-chan []model.Metrics, wg *sync.WaitGroup) {
	zap.L().Info("Starting worker", zap.Int("Worker id", id))
	defer wg.Done()

	for metrics := range jobs {
		err := sender.sendMetrics(metrics)
		if err != nil {
			zap.L().Error("Error sending metrics", zap.Error(err), zap.Int("Worker id", id))
		} else {
			sender.collector.ResetPollCounter()
		}
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

	var encryptedAESKey string
	var encryptedData = compressedBody.Bytes()
	if sender.publicKey != nil {
		encryptedData, encryptedAESKey, err = encryptRequestBody(compressedBody.Bytes(), sender.publicKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt metrics: %w", err)
		}
	}

	request := sender.client.R()

	if sender.key != "" {
		var hash = calculateHash(jsonData, sender.key)
		request.SetHeader("HashSHA256", hash)
	}

	response, err := request.
		SetHeader("X-Real-Ip", sender.localIP.String()).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetHeader("Encrypted-AES-Key", encryptedAESKey).
		SetBody(encryptedData).
		Post("/updates/")

	if err != nil {
		return fmt.Errorf("failed to send metric %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send metric, StatusCode: %d", response.StatusCode())
	}
	return nil
}

func encryptRequestBody(body []byte, publicKey *rsa.PublicKey) ([]byte, string, error) {
	// Generate AES key
	aesKey := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, aesKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Encrypt data with AES key
	encryptedData, err := encryptWithAES(body, aesKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt data with AES: %w", err)
	}

	// Encrypt AES key with RSA public key
	encryptedAESKey, err := encryptWithPublicKey(aesKey, publicKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	return encryptedData, encryptedAESKey, nil
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

// encryptWithPublicKey encrypts data using RSA public key
func encryptWithPublicKey(data []byte, publicKey *rsa.PublicKey) (string, error) {
	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// encryptWithAES encrypts data using AES-GCM
func encryptWithAES(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	encryptedData := aesGCM.Seal(nonce, nonce, data, nil)
	return encryptedData, nil
}

func calculateHash(jsonData []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(jsonData)
	return hex.EncodeToString(hash.Sum(nil))
}

func getLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}
