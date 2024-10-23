package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_storage "github.com/zavtra-na-rabotu/gometrics/internal/mocks"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"

	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
)

func Example() {
	memStorage := storage.NewMemStorage()

	r := chi.NewRouter()

	r.Post("/update", UpdateMetric(memStorage))
	r.Get("/value", GetMetric(memStorage))

	updateMetricRequest := map[string]interface{}{
		"id":    "Counter metric",
		"delta": 12,
		"type":  "counter",
	}
	updateMetricBody, _ := json.Marshal(updateMetricRequest)

	updateMetric := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(updateMetricBody))
	updateMetricRecorder := httptest.NewRecorder()
	r.ServeHTTP(updateMetricRecorder, updateMetric)
	updateMetricResult := updateMetricRecorder.Result()
	fmt.Println(updateMetricResult.StatusCode)

	getMetricRequest := map[string]interface{}{
		"id":   "Counter metric",
		"type": "counter",
	}
	getMetricBody, _ := json.Marshal(getMetricRequest)

	getMetric := httptest.NewRequest(http.MethodGet, "/value", bytes.NewBuffer(getMetricBody))
	getMetricRecorder := httptest.NewRecorder()
	r.ServeHTTP(getMetricRecorder, getMetric)
	getMetricResult := getMetricRecorder.Result()
	fmt.Println(getMetricResult.StatusCode)

	// Output:
	// 200
	// 200
}

func TestUpdateMetric_Common(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name    string
		request map[string]interface{}
		want
	}{
		{
			name: "Negative scenario. Wrong metric type (400)",
			request: map[string]interface{}{
				"id":    "Counter metric",
				"delta": 12,
				"type":  "wrong metric",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Negative scenario. No ID (400)",
			request: map[string]interface{}{
				"delta": 12,
				"type":  "gauge",
			},
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "Negative scenario. Wrong gauge metric field (400)",
			request: map[string]interface{}{
				"id":    "Gauge metric",
				"delta": 12,
				"type":  "gauge",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Negative scenario. Wrong counter metric field (400)",
			request: map[string]interface{}{
				"id":    "Counter metric",
				"value": 12,
				"type":  "counter",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create generated mock
			mockStorage := mock_storage.NewMockStorage(ctrl)

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
			request:       model.Metrics{ID: "Counter metric", Delta: &testDelta, MType: string(model.Counter)},
			storageReturn: storageReturn{12, nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Gauge metric (500)",
			request:       model.Metrics{ID: "Counter metric", Delta: &testDelta, MType: string(model.Counter)},
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
			request:       model.Metrics{ID: "Gauge metric", Value: &testValue, MType: string(model.Gauge)},
			storageReturn: storageReturn{nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Gauge metric (500)",
			request:       model.Metrics{ID: "Gauge metric", Value: &testValue, MType: string(model.Gauge)},
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
			request:       model.Metrics{ID: "Counter metric", Delta: &testDelta, MType: string(model.Counter)},
			storageReturn: storageReturn{testDelta, nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Counter metric not found (404)",
			request:       model.Metrics{ID: "Counter metric", Delta: &testDelta, MType: string(model.Counter)},
			storageReturn: storageReturn{0, storage.ErrItemNotFound},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:          "Negative scenario. Counter metric error (500)",
			request:       model.Metrics{ID: "Counter metric", Delta: &testDelta, MType: string(model.Counter)},
			storageReturn: storageReturn{0, errors.New("some error")},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
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
			request:       model.Metrics{ID: "Gauge metric", Value: &testValue, MType: string(model.Gauge)},
			storageReturn: storageReturn{testValue, nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Gauge metric not found (404)",
			request:       model.Metrics{ID: "Gauge metric", Value: &testValue, MType: string(model.Gauge)},
			storageReturn: storageReturn{0, storage.ErrItemNotFound},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:          "Negative scenario. Gauge metric error (500)",
			request:       model.Metrics{ID: "Gauge metric", Value: &testValue, MType: string(model.Gauge)},
			storageReturn: storageReturn{0, errors.New("some error")},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
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
