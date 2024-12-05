// Package v3 contains handlers for API version 3
package v3

import (
	"encoding/json"
	"net/http"

	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"go.uber.org/zap"
)

// Ping verifies a connection to the database is still alive
func Ping(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := st.(*storage.DBStorage); !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := st.(*storage.DBStorage).Ping()
		if err != nil {
			zap.L().Error("Database ping failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// UpdateMetrics handler to update batch of metrics
func UpdateMetrics(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []model.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			zap.L().Error("Failed to read body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := st.UpdateMetrics(metrics)
		if err != nil {
			zap.L().Error("Failed to update metrics", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
