package handlers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateMetricHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	type request struct {
		metricType  string
		metricName  string
		metricValue string
	}

	tests := []struct {
		name    string
		arg     storage.Storage
		request request
		want    want
	}{
		{
			name: "Positive scenario. Gauge metric (200)",
			arg:  storage.NewMemStorage(),
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
			request: request{
				metricType:  "gauge",
				metricName:  "test",
				metricValue: "28.0",
			},
		},
		{
			name: "Negative scenario. No type provided (400)",
			arg:  storage.NewMemStorage(),
			want: want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
			request: request{
				metricType:  "wrong",
				metricName:  "test",
				metricValue: "28.0",
			},
		},
		{
			name: "Negative scenario. No name provided (404)",
			arg:  storage.NewMemStorage(),
			want: want{
				contentType: "",
				statusCode:  http.StatusNotFound,
			},
			request: request{
				metricType:  "gauge",
				metricName:  "",
				metricValue: "28.0",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := fmt.Sprintf("/update/%s/%s/%s", test.request.metricType, test.request.metricName, test.request.metricValue)

			request := httptest.NewRequest(http.MethodPost, url, nil)
			request.SetPathValue("type", test.request.metricType)
			request.SetPathValue("name", test.request.metricName)
			request.SetPathValue("value", test.request.metricValue)

			responseRecorder := httptest.NewRecorder()
			handler := UpdateMetricHandler(test.arg)
			handler(responseRecorder, request)

			result := responseRecorder.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, test.want.statusCode, result.StatusCode)
			assert.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
