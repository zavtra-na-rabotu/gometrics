package v1

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_storage "github.com/zavtra-na-rabotu/gometrics/internal/mocks"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
)

func Example() {
	// Change to project root and back after test finish
	originalDir, _ := os.Getwd()

	// TODO: Dont know how to fix it
	os.Chdir("../../../..")
	defer os.Chdir(originalDir)

	memStorage := storage.NewMemStorage()

	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", UpdateMetric(memStorage))
	r.Get("/value/{type}/{name}", GetMetric(memStorage))
	r.Get("/", RenderAllMetrics(memStorage))

	// Пример запроса и ответа:
	createMetric := httptest.NewRequest(http.MethodPost, "/update/gauge/cpu/10", nil)
	createMetricRecorder := httptest.NewRecorder()
	r.ServeHTTP(createMetricRecorder, createMetric)
	createMetricResult := createMetricRecorder.Result()
	fmt.Println(createMetricResult.StatusCode)

	getMetric := httptest.NewRequest(http.MethodGet, "/value/gauge/cpu", nil)
	getMetricRecorder := httptest.NewRecorder()
	r.ServeHTTP(getMetricRecorder, getMetric)
	getMetricResult := getMetricRecorder.Result()
	fmt.Println(getMetricResult.StatusCode)

	getAllMetrics := httptest.NewRequest(http.MethodGet, "/", nil)
	getAllMetricsRecorder := httptest.NewRecorder()
	r.ServeHTTP(getAllMetricsRecorder, getAllMetrics)
	getAllMetricsResult := getAllMetricsRecorder.Result()
	fmt.Println(getAllMetricsResult.StatusCode)

	// Output:
	// 200
	// 200
	// 200
}

func TestUpdateMetric_Common(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type storageReturn struct {
		err error
	}

	testValue := int64(12)

	tests := []struct {
		name          string
		request       model.Metrics
		storageReturn storageReturn
		want
	}{
		{
			name:          "No name provided",
			request:       model.Metrics{ID: "", Delta: &testValue, MType: string(model.Counter)},
			storageReturn: storageReturn{nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:          "No type provided",
			request:       model.Metrics{ID: "Whatever metric", Delta: &testValue, MType: ""},
			storageReturn: storageReturn{errors.New("some error")},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create generated mock
			mockStorage := mock_storage.NewMockStorage(ctrl)

			// Create request
			request := httptest.NewRequest(http.MethodPost, "/update/", http.NoBody)
			request.SetPathValue("type", test.request.MType)
			request.SetPathValue("name", test.request.ID)
			request.SetPathValue("value", strconv.FormatInt(*test.request.Delta, 10))

			responseRecorder := httptest.NewRecorder()

			handler := UpdateMetric(mockStorage)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()

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
		err error
	}

	testValue := int64(12)

	tests := []struct {
		name          string
		request       model.Metrics
		storageReturn storageReturn
		want
	}{
		{
			name:          "Positive scenario. Counter metric (200)",
			request:       model.Metrics{ID: "Counter metric", Delta: &testValue, MType: string(model.Counter)},
			storageReturn: storageReturn{nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Counter metric (500)",
			request:       model.Metrics{ID: "Counter metric", Delta: &testValue, MType: string(model.Counter)},
			storageReturn: storageReturn{errors.New("some error")},
			want: want{
				contentType: "",
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
			mockStorage.EXPECT().UpdateCounter(test.request.ID, *test.request.Delta).Return(test.storageReturn.err)

			// Create request
			request := httptest.NewRequest(http.MethodPost, "/update/", http.NoBody)
			request.SetPathValue("type", test.request.MType)
			request.SetPathValue("name", test.request.ID)
			request.SetPathValue("value", strconv.FormatInt(*test.request.Delta, 10))

			responseRecorder := httptest.NewRecorder()

			handler := UpdateMetric(mockStorage)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
			assert.Equal(t, test.want.contentType, responseRecorder.Header().Get("Content-Type"))
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
				contentType: "",
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

			// Create request
			request := httptest.NewRequest(http.MethodPost, "/update/", http.NoBody)
			request.SetPathValue("type", test.request.MType)
			request.SetPathValue("name", test.request.ID)
			request.SetPathValue("value", strconv.FormatFloat(*test.request.Value, 'f', -1, 64))

			responseRecorder := httptest.NewRecorder()

			handler := UpdateMetric(mockStorage)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
			assert.Equal(t, test.want.contentType, responseRecorder.Header().Get("Content-Type"))
		})
	}
}

func TestGetMetric_Common(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name    string
		request model.Metrics
		want    want
	}{
		{
			name:    "Negative scenario. Empty name",
			request: model.Metrics{MType: string(model.Counter)},
			want: want{
				contentType: "",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:    "Negative scenario. Empty type",
			request: model.Metrics{ID: "whatever"},
			want: want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "Negative scenario. Wrong type",
			request: model.Metrics{ID: "whatever", MType: "whatever"},
			want: want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create generated mock
			mockStorage := mock_storage.NewMockStorage(ctrl)

			// Create request
			request := httptest.NewRequest(http.MethodGet, "/value/", http.NoBody)
			request.SetPathValue("type", test.request.MType)
			request.SetPathValue("name", test.request.ID)

			responseRecorder := httptest.NewRecorder()

			handler := GetMetric(mockStorage)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
			assert.Equal(t, test.want.contentType, responseRecorder.Header().Get("Content-Type"))
		})
	}
}

func TestGetMetric_Counter(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type storageReturn struct {
		value int64
		err   error
	}

	tests := []struct {
		name          string
		request       model.Metrics
		storageReturn storageReturn
		want          want
	}{
		{
			name:          "Positive scenario. Counter metric (200)",
			request:       model.Metrics{ID: "whatever", MType: string(model.Counter)},
			storageReturn: storageReturn{15, nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Counter metric (400)",
			request:       model.Metrics{ID: "whatever", MType: string(model.Counter)},
			storageReturn: storageReturn{0, storage.ErrItemNotFound},
			want: want{
				contentType: "",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:          "Negative scenario. Counter metric (500)",
			request:       model.Metrics{ID: "whatever", MType: string(model.Counter)},
			storageReturn: storageReturn{0, errors.New("some error")},
			want: want{
				contentType: "",
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
			mockStorage.EXPECT().GetCounter(test.request.ID).Return(test.storageReturn.value, test.storageReturn.err)

			// Create request
			request := httptest.NewRequest(http.MethodGet, "/value/", http.NoBody)
			request.SetPathValue("type", test.request.MType)
			request.SetPathValue("name", test.request.ID)

			responseRecorder := httptest.NewRecorder()

			handler := GetMetric(mockStorage)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
			assert.Equal(t, test.want.contentType, responseRecorder.Header().Get("Content-Type"))
		})
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

	tests := []struct {
		name          string
		request       model.Metrics
		storageReturn storageReturn
		want          want
	}{
		{
			name:          "Positive scenario. Gauge metric (200)",
			request:       model.Metrics{ID: "whatever", MType: string(model.Gauge)},
			storageReturn: storageReturn{15.4, nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Gauge metric (400)",
			request:       model.Metrics{ID: "whatever", MType: string(model.Gauge)},
			storageReturn: storageReturn{0, storage.ErrItemNotFound},
			want: want{
				contentType: "",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:          "Negative scenario. Gauge metric (500)",
			request:       model.Metrics{ID: "whatever", MType: string(model.Gauge)},
			storageReturn: storageReturn{0, errors.New("some error")},
			want: want{
				contentType: "",
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
			mockStorage.EXPECT().GetGauge(test.request.ID).Return(test.storageReturn.value, test.storageReturn.err)

			// Create request
			request := httptest.NewRequest(http.MethodGet, "/value/", http.NoBody)
			request.SetPathValue("type", test.request.MType)
			request.SetPathValue("name", test.request.ID)

			responseRecorder := httptest.NewRecorder()

			handler := GetMetric(mockStorage)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
			assert.Equal(t, test.want.contentType, responseRecorder.Header().Get("Content-Type"))
		})
	}
}

func TestRenderAllMetrics_Common(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type gaugeStorageReturn struct {
		value map[string]float64
		err   error
	}

	type counterStorageReturn struct {
		value map[string]int64
		err   error
	}

	tests := []struct {
		name                 string
		gaugeStorageReturn   gaugeStorageReturn
		counterStorageReturn counterStorageReturn
		want                 want
	}{
		{
			name: "Positive scenario (200)",
			counterStorageReturn: counterStorageReturn{map[string]int64{
				"cpu":  123,
				"heap": 321,
			}, nil},
			gaugeStorageReturn: gaugeStorageReturn{map[string]float64{
				"whatever": 12.5,
				"bla":      1.5,
			}, nil},
			want: want{
				contentType: "text/html; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Change to project root and back after test finish
			originalDir, _ := os.Getwd()

			// TODO: Dont know how to fix it
			os.Chdir("../../../..")
			defer os.Chdir(originalDir)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create generated mock
			mockStorage := mock_storage.NewMockStorage(ctrl)
			mockStorage.EXPECT().GetAllGauge().Return(test.gaugeStorageReturn.value, test.gaugeStorageReturn.err)
			mockStorage.EXPECT().GetAllCounter().Return(test.counterStorageReturn.value, test.counterStorageReturn.err)

			// Create request
			request := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

			responseRecorder := httptest.NewRecorder()

			handler := RenderAllMetrics(mockStorage)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
			assert.Equal(t, test.want.contentType, responseRecorder.Header().Get("Content-Type"))
		})
	}
}

func TestRenderAllMetrics_GaugeError(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type gaugeStorageReturn struct {
		value map[string]float64
		err   error
	}

	type counterStorageReturn struct {
		value map[string]int64
		err   error
	}

	tests := []struct {
		name                 string
		gaugeStorageReturn   gaugeStorageReturn
		counterStorageReturn counterStorageReturn
		want                 want
	}{
		{
			name: "Gauge negative scenario (500)",
			counterStorageReturn: counterStorageReturn{map[string]int64{
				"cpu":  123,
				"heap": 321,
			}, nil},
			gaugeStorageReturn: gaugeStorageReturn{nil, errors.New("some error")},
			want: want{
				contentType: "text/html; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Change to project root and back after test finish
			originalDir, _ := os.Getwd()

			// TODO: Dont know how to fix it
			os.Chdir("../../../..")
			defer os.Chdir(originalDir)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create generated mock
			mockStorage := mock_storage.NewMockStorage(ctrl)
			mockStorage.EXPECT().GetAllGauge().Return(test.gaugeStorageReturn.value, test.gaugeStorageReturn.err)

			// Create request
			request := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

			responseRecorder := httptest.NewRecorder()

			handler := RenderAllMetrics(mockStorage)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
			assert.Equal(t, test.want.contentType, responseRecorder.Header().Get("Content-Type"))
		})
	}
}

func TestRenderAllMetrics_CounterError(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type gaugeStorageReturn struct {
		value map[string]float64
		err   error
	}

	type counterStorageReturn struct {
		value map[string]int64
		err   error
	}

	tests := []struct {
		name                 string
		gaugeStorageReturn   gaugeStorageReturn
		counterStorageReturn counterStorageReturn
		want                 want
	}{
		{
			name:                 "Counter negative scenario (500)",
			counterStorageReturn: counterStorageReturn{nil, errors.New("some error")},
			gaugeStorageReturn: gaugeStorageReturn{map[string]float64{
				"whatever": 12.5,
				"bla":      1.5,
			}, nil},
			want: want{
				contentType: "text/html; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Change to project root and back after test finish
			originalDir, _ := os.Getwd()

			// TODO: Dont know how to fix it
			os.Chdir("../../../..")
			defer os.Chdir(originalDir)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create generated mock
			mockStorage := mock_storage.NewMockStorage(ctrl)
			mockStorage.EXPECT().GetAllGauge().Return(test.gaugeStorageReturn.value, test.gaugeStorageReturn.err)
			mockStorage.EXPECT().GetAllCounter().Return(test.counterStorageReturn.value, test.counterStorageReturn.err)

			// Create request
			request := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

			responseRecorder := httptest.NewRecorder()

			handler := RenderAllMetrics(mockStorage)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
			assert.Equal(t, test.want.contentType, responseRecorder.Header().Get("Content-Type"))
		})
	}
}
