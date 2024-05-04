package handlers

import (
	"github.com/zavtra-na-rabotu/gometrics/internal"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"log"
	"net/http"
	"strconv"
)

func UpdateMetricHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate request method
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Get data from request
		metricType := internal.MetricType(r.PathValue("type"))
		metricName := r.PathValue("name")
		metricValue := r.PathValue("value")

		// Validate metricType
		if !(metricType == internal.Counter || metricType == internal.Gauge) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate metricName
		if stringutils.IsEmpty(metricName) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch metricType {
		case internal.Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				log.Printf("ERROR: failed to parse value from metric '%s': %s\n", metricValue, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(metricName, value)
		case internal.Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				log.Printf("ERROR: failed to parse value from metric '%s': %s\n", metricValue, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(metricName, value)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}
