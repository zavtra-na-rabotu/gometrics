package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGzipCompression(t *testing.T) {
	requestBody := `{
		"id": "WhateverMetric",
		"type": "gauge",
		"value": "5"
	}`

	successBody := `{
		"id": "WhateverMetric",
		"type": "gauge",
		"value": 5
	}`

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(successBody))
	})

	middleware := GzipMiddleware(handler)

	server := httptest.NewServer(middleware)
	defer server.Close()

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)

		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)

		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, server.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		}()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest(http.MethodPost, server.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		}()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}
