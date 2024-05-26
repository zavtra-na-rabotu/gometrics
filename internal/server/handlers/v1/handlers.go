package v1

import (
	"errors"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"go.uber.org/zap"
)

type MetricResponse struct {
	MetricType  model.MetricType
	MetricName  string
	MetricValue string
}

// GetMetric TODO: What to do if variable name and package name are the same ? Only aliases ?
func GetMetric(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := model.MetricType(r.PathValue("type"))
		metricName := r.PathValue("name")

		// Validate MetricType TODO: Remove duplicates
		if !(metricType == model.Counter || metricType == model.Gauge) {
			zap.L().Error("Invalid metric type", zap.String("metricType", string(metricType)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate MetricName TODO: Remove duplicates
		if stringutils.IsEmpty(metricName) {
			zap.L().Error("Invalid metric name", zap.String("metricName", metricName))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var response string

		switch metricType {
		case model.Counter:
			metric, err := st.GetCounter(metricName)
			if err != nil {
				if errors.Is(err, storage.ErrItemNotFound) {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				zap.L().Error("Error while getting counter metric", zap.String("metricName", metricName), zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			response = strconv.FormatInt(metric, 10)
		case model.Gauge:
			metric, err := st.GetGauge(metricName)
			if err != nil {
				if errors.Is(err, storage.ErrItemNotFound) {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				zap.L().Error("Error while getting gauge metric", zap.String("metricName", metricName), zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			response = strconv.FormatFloat(metric, 'f', -1, 64)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := w.Write([]byte(response))
		if err != nil {
			zap.L().Error("Error while writing response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func UpdateMetric(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get data from request
		metricType := model.MetricType(r.PathValue("type"))
		metricName := r.PathValue("name")
		metricValue := r.PathValue("value")

		// Validate MetricType TODO: Remove duplicates
		if !(metricType == model.Counter || metricType == model.Gauge) {
			zap.L().Error("Invalid metric type", zap.String("metricType", string(metricType)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate MetricName TODO: Remove duplicates
		if stringutils.IsEmpty(metricName) {
			zap.L().Error("Invalid metric name", zap.String("metricName", metricName))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch metricType {
		case model.Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				zap.L().Error("Failed to parse value from metric", zap.String("metricValue", metricValue), zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			st.UpdateCounter(metricName, value)
		case model.Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				zap.L().Error("Failed to parse value from metric", zap.String("metricValue", metricValue), zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			st.UpdateGauge(metricName, value)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

func RenderAllMetrics(st *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wd, _ := os.Getwd()
		metricsTemplate, err := template.ParseFiles(wd + "/internal/server/web/metrics/metrics.tmpl")
		if err != nil {
			zap.L().Error("Error while parsing template", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		allMetrics := []MetricResponse{}

		for name, metric := range st.GetAllGauge() {
			allMetrics = append(allMetrics, MetricResponse{
				MetricType:  model.Gauge,
				MetricName:  name,
				MetricValue: strconv.FormatFloat(metric, 'f', -1, 64),
			})
		}

		for name, metric := range st.GetAllCounter() {
			allMetrics = append(allMetrics, MetricResponse{
				MetricType:  model.Counter,
				MetricName:  name,
				MetricValue: strconv.FormatInt(metric, 10),
			})
		}

		err = metricsTemplate.Execute(w, allMetrics)
		if err != nil {
			zap.L().Error("Error while executing template", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
}