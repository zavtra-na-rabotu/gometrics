package v2

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/zavtra-na-rabotu/gometrics/internal/mocks"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/middleware"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
)

func TestGzipCompression(t *testing.T) {
	handler := middleware.GzipMiddleware(UpdateMetric(storage.NewMemStorage()))

	srv := httptest.NewServer(handler)

	defer srv.Close()

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

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, srv.URL, buf)
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
		r := httptest.NewRequest(http.MethodPost, srv.URL, buf)
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

func TestUpdateMetric_Counter(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type storageReturn struct {
		counter int64
		err     error
	}

	testDelta := int64(12)

	tests := []struct {
		name          string
		request       model.Metrics
		storageReturn storageReturn
		want
	}{
		{
			name:          "Positive scenario. Counter metric (200)",
			request:       model.Metrics{ID: "Counter metric", Delta: &testDelta, MType: "counter"},
			storageReturn: storageReturn{12, nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Gauge metric (500)",
			request:       model.Metrics{ID: "Counter metric", Delta: &testDelta, MType: "counter"},
			storageReturn: storageReturn{0, errors.New("some error")},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create generated mock
			mockStorage := mock_storage.NewMockStorage(ctrl)
			mockStorage.EXPECT().UpdateCounterAndReturn(test.request.ID, *test.request.Delta).Return(test.storageReturn.counter, test.storageReturn.err)

			// Metrics to JSON
			metricsJSON, _ := json.Marshal(test.request)

			// Create request
			request, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(metricsJSON))
			if err != nil {
				t.Fatal(err)
			}

			responseRecorder := httptest.NewRecorder()

			handler := UpdateMetric(mockStorage)
			handler.ServeHTTP(responseRecorder, request)

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
		})
	}
}

func TestUpdateMetric_Gauge(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type storageReturn struct {
		err error
	}

	testValue := 23.4

	tests := []struct {
		name          string
		request       model.Metrics
		storageReturn storageReturn
		want
	}{
		{
			name:          "Positive scenario. Gauge metric (200)",
			request:       model.Metrics{ID: "Gauge metric", Value: &testValue, MType: "gauge"},
			storageReturn: storageReturn{nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Gauge metric (500)",
			request:       model.Metrics{ID: "Gauge metric", Value: &testValue, MType: "gauge"},
			storageReturn: storageReturn{errors.New("some error")},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create generated mock
			mockStorage := mock_storage.NewMockStorage(ctrl)
			mockStorage.EXPECT().UpdateGauge(test.request.ID, *test.request.Value).Return(test.storageReturn.err)

			// Metrics to JSON
			metricsJSON, _ := json.Marshal(test.request)

			// Create request
			request, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(metricsJSON))
			if err != nil {
				t.Fatal(err)
			}

			responseRecorder := httptest.NewRecorder()

			handler := UpdateMetric(mockStorage)
			handler.ServeHTTP(responseRecorder, request)

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
		})
	}
}

func TestGetMetric_Counter(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type storageReturn struct {
		delta int64
		err   error
	}

	testDelta := int64(12)

	tests := []struct {
		name          string
		request       model.Metrics
		storageReturn storageReturn
		want
	}{
		{
			name:          "Positive scenario. Counter metric (200)",
			request:       model.Metrics{ID: "Counter metric", Delta: &testDelta, MType: "counter"},
			storageReturn: storageReturn{testDelta, nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
	}
	for _, test := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create generated mock
		mockStorage := mock_storage.NewMockStorage(ctrl)
		mockStorage.EXPECT().GetCounter(test.request.ID).Return(test.storageReturn.delta, test.storageReturn.err)

		// Metrics to JSON
		metricsJSON, _ := json.Marshal(test.request)

		// Create request
		request, err := http.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(metricsJSON))
		if err != nil {
			t.Fatal(err)
		}

		responseRecorder := httptest.NewRecorder()

		handler := GetMetric(mockStorage)
		handler.ServeHTTP(responseRecorder, request)

		assert.NoError(t, err)
		assert.Equal(t, test.want.statusCode, responseRecorder.Code)
	}
}

func TestGetMetric_Gauge(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type storageReturn struct {
		value float64
		err   error
	}

	testValue := 23.4

	tests := []struct {
		name          string
		request       model.Metrics
		storageReturn storageReturn
		want
	}{
		{
			name:          "Positive scenario. Gauge metric (200)",
			request:       model.Metrics{ID: "Gauge metric", Value: &testValue, MType: "gauge"},
			storageReturn: storageReturn{testValue, nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
	}
	for _, test := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create generated mock
		mockStorage := mock_storage.NewMockStorage(ctrl)
		mockStorage.EXPECT().GetGauge(test.request.ID).Return(test.storageReturn.value, test.storageReturn.err)

		// Metrics to JSON
		metricsJSON, _ := json.Marshal(test.request)

		// Create request
		request, err := http.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(metricsJSON))
		if err != nil {
			t.Fatal(err)
		}

		responseRecorder := httptest.NewRecorder()

		handler := GetMetric(mockStorage)
		handler.ServeHTTP(responseRecorder, request)

		assert.NoError(t, err)
		assert.Equal(t, test.want.statusCode, responseRecorder.Code)
	}
}
