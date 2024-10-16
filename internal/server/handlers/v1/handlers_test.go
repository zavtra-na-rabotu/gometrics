package v1

import (
	"errors"
	"github.com/golang/mock/gomock"
	mock_storage "github.com/zavtra-na-rabotu/gometrics/internal/mocks"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			request:       model.Metrics{ID: "", Delta: &testValue, MType: "counter"},
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
			request:       model.Metrics{ID: "Counter metric", Delta: &testValue, MType: "counter"},
			storageReturn: storageReturn{nil},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:          "Negative scenario. Counter metric (500)",
			request:       model.Metrics{ID: "Counter metric", Delta: &testValue, MType: "counter"},
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
		})
	}
}
