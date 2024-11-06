package metrics

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
)

func TestNewSender(t *testing.T) {
	url := "localhost:8080"
	key := "testKey"
	rateLimit := 2
	reportInterval := 5

	collector := &Collector{}

	sender := NewSender(url, key, rateLimit, reportInterval, collector)

	assert.NotNil(t, sender)
	assert.Equal(t, "http://"+url, sender.client.BaseURL)
	assert.Equal(t, key, sender.key)
	assert.Equal(t, rateLimit, sender.rateLimit)
	assert.Equal(t, time.Duration(reportInterval)*time.Second, sender.reportInterval)
}

func TestSender_SendMetrics(t *testing.T) {
	value := 1.23
	hashKey := "testKey"

	metrics := []model.Metrics{
		{ID: "testMetric", MType: "gauge", Value: &value},
	}

	jsonData, err := json.Marshal(metrics)
	assert.NoError(t, err)

	expectedHash := calculateHash(jsonData, hashKey)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		receivedHash := r.Header.Get("HashSHA256")
		assert.Equal(t, expectedHash, receivedHash)

		var compressedBody bytes.Buffer
		_, err := compressedBody.ReadFrom(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		gzipReader, err := gzip.NewReader(&compressedBody)
		assert.NoError(t, err)
		defer gzipReader.Close()

		body, err := io.ReadAll(gzipReader)
		assert.NoError(t, err)

		var receivedMetrics []model.Metrics
		err = json.Unmarshal(body, &receivedMetrics)
		assert.NoError(t, err)
		assert.Equal(t, metrics, receivedMetrics)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	sender := &Sender{client: client, key: hashKey}

	err = sender.sendMetrics(metrics)
	assert.NoError(t, err)
}

func TestSender_CalculateHash(t *testing.T) {
	data := []byte("test data")
	key := "testkey"

	expectedHash := hmac.New(sha256.New, []byte(key))
	expectedHash.Write(data)

	expected := hex.EncodeToString(expectedHash.Sum(nil))
	actual := calculateHash(data, key)

	assert.Equal(t, expected, actual)
}

func TestSender_CompressBody(t *testing.T) {
	data := []byte("test data")
	var compressedData bytes.Buffer

	err := compressBody(&compressedData, data)
	assert.NoError(t, err)

	gzipReader, err := gzip.NewReader(&compressedData)
	assert.NoError(t, err)
	defer gzipReader.Close()

	decompressedData, err := io.ReadAll(gzipReader)
	assert.NoError(t, err)
	assert.Equal(t, data, decompressedData)
}
