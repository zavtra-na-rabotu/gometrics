package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_UnmarshalJSON_Positive(t *testing.T) {
	type fields struct {
		Value *float64
		Delta *int64
		ID    string
		MType string
	}
	type args struct {
		json []byte
	}

	delta := int64(15)
	value := 15.5

	tests := []struct {
		name string
		want fields
		args args
	}{
		{
			name: "Should unmarshal JSON. Counter metric as integer",
			want: fields{
				ID:    "whatever",
				MType: "counter",
				Delta: &delta,
			},
			args: args{
				json: []byte(`{"id": "whatever", "type": "counter", "delta": 15}`),
			},
		},
		{
			name: "Should unmarshal JSON. Counter metric as string",
			want: fields{
				ID:    "whatever",
				MType: "counter",
				Delta: &delta,
			},
			args: args{
				json: []byte(`{"id": "whatever", "type": "counter", "delta": "15"}`),
			},
		},
		{
			name: "Should unmarshal JSON. Gauge metric as float",
			want: fields{
				ID:    "whatever",
				MType: "gauge",
				Value: &value,
			},
			args: args{
				json: []byte(`{"id": "whatever", "type": "gauge", "value": 15.5}`),
			},
		},
		{
			name: "Should unmarshal JSON. Gauge metric as string",
			want: fields{
				ID:    "whatever",
				MType: "gauge",
				Value: &value,
			},
			args: args{
				json: []byte(`{"id": "whatever", "type": "gauge", "value": "15.5"}`),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := &Metrics{}

			err := metric.UnmarshalJSON(test.args.json)

			assert.NoError(t, err)

			assert.Equal(t, test.want.ID, metric.ID)
			assert.Equal(t, test.want.MType, metric.MType)
			assert.Equal(t, test.want.Delta, metric.Delta)
			assert.Equal(t, test.want.Value, metric.Value)
		})
	}
}

func TestMetrics_UnmarshalJSON_Negative(t *testing.T) {
	type args struct {
		json []byte
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Should not unmarshal JSON. Counter metric as wrong string",
			args: args{
				json: []byte(`{"id": "whatever", "type": "counter", "delta": "test"}`),
			},
		},
		{
			name: "Should not unmarshal JSON. Gauge metric as wrong string",
			args: args{
				json: []byte(`{"id": "whatever", "type": "gauge", "value": "test"}`),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := &Metrics{}

			err := metric.UnmarshalJSON(test.args.json)

			assert.NotNil(t, err)
		})
	}
}
