package v2

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"go.uber.org/zap"
)

func UpdateMetric(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics model.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			zap.L().Error("Failed to read body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !(metrics.MType == string(model.Counter) || metrics.MType == string(model.Gauge)) {
			zap.L().Error("Invalid metric type", zap.String("type", metrics.MType))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if stringutils.IsEmpty(metrics.ID) {
			zap.L().Error("Invalid metric name", zap.String("name", metrics.ID))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch metrics.MType {
		case string(model.Counter):
			*metrics.Delta = st.UpdateCounterAndReturn(metrics.ID, *metrics.Delta)
		case string(model.Gauge):
			st.UpdateGauge(metrics.ID, *metrics.Value)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&metrics); err != nil {
			zap.L().Error("Failed to write response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func GetMetric(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics model.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			zap.L().Error("Failed to read body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !(metrics.MType == string(model.Counter) || metrics.MType == string(model.Gauge)) {
			zap.L().Error("Invalid metric type", zap.String("type", metrics.MType))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if stringutils.IsEmpty(metrics.ID) {
			zap.L().Error("Invalid metric name", zap.String("name", metrics.ID))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch metrics.MType {
		case string(model.Counter):
			delta, err := st.GetCounter(metrics.ID)
			if err != nil {
				if errors.Is(err, storage.ErrItemNotFound) {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				zap.L().Error("Error while getting counter metric", zap.String("metricName", metrics.ID), zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			metrics.Delta = &delta
		case string(model.Gauge):
			value, err := st.GetGauge(metrics.ID)
			if err != nil {
				if errors.Is(err, storage.ErrItemNotFound) {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				zap.L().Error("Error while getting gauge metric", zap.String("metricName", metrics.ID), zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			metrics.Value = &value
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&metrics); err != nil {
			zap.L().Error("Failed to write response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}