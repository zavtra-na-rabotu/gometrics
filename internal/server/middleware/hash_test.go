package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseHashMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	})

	key := "secret-key"
	middleware := ResponseHashMiddleware(key)

	server := httptest.NewServer(middleware(handler))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("could not send GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}

	expectedHash := calculateHash(body, key)
	receivedHash := resp.Header.Get("HashSHA256")
	assert.Equal(t, expectedHash, receivedHash, "Response hash does not match")
}

func TestRequestHashMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	})

	key := "secret-key"
	middleware := RequestHashMiddleware(key)

	server := httptest.NewServer(middleware(handler))
	defer server.Close()

	requestBody := `{
		"id": "WhateverMetric",
		"type": "gauge",
		"value": "5"
	}`
	body := []byte(requestBody)
	calculatedHash := calculateHash(body, key)
	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("could not create POST request: %v", err)
	}
	req.Header.Set("HashSHA256", calculatedHash)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("could not send POST request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestHashMiddleware_InvalidHash(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	})

	key := "secret-key"
	middleware := RequestHashMiddleware(key)

	server := httptest.NewServer(middleware(handler))
	defer server.Close()

	requestBody := `{
		"id": "WhateverMetric",
		"type": "gauge",
		"value": "5"
	}`
	body := []byte(requestBody)
	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("could not create POST request: %v", err)
	}
	req.Header.Set("HashSHA256", "invalid-hash")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("could not send POST request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
