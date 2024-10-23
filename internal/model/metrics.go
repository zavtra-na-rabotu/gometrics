package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Metrics main structure to store all types of metrics
type Metrics struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}

// UnmarshalJSON custom logic for unmarshalling JSON to Metrics structure
func (m *Metrics) UnmarshalJSON(data []byte) (err error) {
	type MetricsAlias Metrics

	aliasValue := &struct {
		*MetricsAlias
		Delta any `json:"delta,omitempty"`
		Value any `json:"value,omitempty"`
	}{
		MetricsAlias: (*MetricsAlias)(m),
	}

	err = json.Unmarshal(data, aliasValue)
	if err != nil {
		return fmt.Errorf("unmarshal metrics error: %w", err)
	}

	if aliasValue.Delta != nil {
		switch v := aliasValue.Delta.(type) {
		case float64:
			delta := int64(v)
			m.Delta = &delta
		case string:
			delta, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse string: %s to int %w", v, err)
			}
			m.Delta = &delta
		default:
			return fmt.Errorf("unexpected type for delta: %T", aliasValue.Delta)
		}
	}

	if aliasValue.Value != nil {
		switch v := aliasValue.Value.(type) {
		case float64:
			m.Value = &v
		case string:
			value, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("failed to parse string: %s to float %w", v, err)
			}
			m.Value = &value
		default:
			return fmt.Errorf("unexpected type for value: %T", aliasValue.Value)
		}
	}

	return nil
}
