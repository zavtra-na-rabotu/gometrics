package v3

import (
	"net/http"

	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"go.uber.org/zap"
)

func Ping(storage *storage.DBStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := storage.Ping()
		if err != nil {
			zap.L().Error("Database ping failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
