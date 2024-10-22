package v3

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_storage "github.com/zavtra-na-rabotu/gometrics/internal/mocks"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
)

func TestUpdateMetrics(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	testValue := 23.4
	testDelta := int64(12)

	tests := []struct {
		name           string
		request        []model.Metrics
		storageReturns interface{}
		want
	}{
		{
			name: "Positive scenario. Two metrics (200)",
			request: []model.Metrics{
				{ID: "Gauge metric", Value: &testValue, MType: "gauge"},
				{ID: "Value metric", Delta: &testDelta, MType: "delta"},
			},
			storageReturns: nil,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:           "Positive scenario. No metrics (200)",
			request:        []model.Metrics{},
			storageReturns: nil,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:           "Negative scenario. Gauge metric (500)",
			request:        []model.Metrics{},
			storageReturns: errors.New("something went wrong"),
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
			mockStorage.EXPECT().UpdateMetrics(test.request).Return(test.storageReturns)

			// Metrics to JSON
			metricsJSON, _ := json.Marshal(test.request)

			// Create request
			request, err := http.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(metricsJSON))
			if err != nil {
				t.Fatal(err)
			}

			responseRecorder := httptest.NewRecorder()

			handler := UpdateMetrics(mockStorage)
			handler.ServeHTTP(responseRecorder, request)

			assert.Equal(t, test.want.statusCode, responseRecorder.Code)
			assert.NoError(t, err)
		})
	}
}
