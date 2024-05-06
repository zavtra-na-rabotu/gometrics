package handlers

import (
	"errors"
	"fmt"
	"github.com/zavtra-na-rabotu/gometrics/internal"
	"github.com/zavtra-na-rabotu/gometrics/internal/server/storage"
	"github.com/zavtra-na-rabotu/gometrics/internal/utils/stringutils"
	"html/template"
	"net/http"
	"os"
	"strconv"
)

type MetricResponse struct {
	MetricType  internal.MetricType
	MetricName  string
	MetricValue any // TODO: Or cast to string ?
}

// TODO: What to do if variable name and package name are the same ? Only aliases ?
func GetMetric(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := internal.MetricType(r.PathValue("type"))
		metricName := r.PathValue("name")

		// Validate MetricType TODO: Remove duplicates
		if !(metricType == internal.Counter || metricType == internal.Gauge) {
			internal.ErrorLog.Println("Invalid metric type: %s", metricType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate MetricName TODO: Remove duplicates
		if stringutils.IsEmpty(metricName) {
			internal.ErrorLog.Println("Invalid metric name: %s", metricName)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var response string

		switch metricType {
		case internal.Counter:
			metric, err := st.GetCounter(metricName)
			if err != nil {
				if errors.Is(err, storage.ErrorItemNotFound) {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				internal.ErrorLog.Println("Error while getting counter metric: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			response = fmt.Sprintf("%d", metric)
		case internal.Gauge:
			metric, err := st.GetGauge(metricName)
			if err != nil {
				if errors.Is(err, storage.ErrorItemNotFound) {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				internal.ErrorLog.Println("Error while getting gauge metric: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			response = fmt.Sprintf("%f", metric)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := w.Write([]byte(response))
		if err != nil {
			internal.ErrorLog.Println("Error writing response:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func UpdateMetric(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get data from request
		metricType := internal.MetricType(r.PathValue("type"))
		metricName := r.PathValue("name")
		metricValue := r.PathValue("value")

		// Validate MetricType TODO: Remove duplicates
		if !(metricType == internal.Counter || metricType == internal.Gauge) {
			internal.ErrorLog.Println("Invalid metric type: %s", metricType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate MetricName TODO: Remove duplicates
		if stringutils.IsEmpty(metricName) {
			internal.ErrorLog.Println("Invalid metric name: %s", metricName)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch metricType {
		case internal.Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				internal.ErrorLog.Println("Failed to parse value from metric '%s': %s\n", metricValue, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(metricName, value)
		case internal.Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				internal.ErrorLog.Println("Failed to parse value from metric '%s': %s\n", metricValue, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(metricName, value)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

func RenderAllMetrics(storage *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wd, _ := os.Getwd()
		metricsTemplate, err := template.ParseFiles(wd + "/internal/server/web/metrics/metrics.tmpl")
		if err != nil {
			internal.ErrorLog.Println("Error while parsing template: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		allMetrics := []MetricResponse{}

		//gaugeMetrics := storage.GetAllGauge()
		for name, metric := range storage.GetAllGauge() {
			allMetrics = append(allMetrics, MetricResponse{
				MetricType:  internal.Gauge,
				MetricName:  name,
				MetricValue: metric,
			})
		}

		for name, metric := range storage.GetAllCounter() {
			allMetrics = append(allMetrics, MetricResponse{
				MetricType:  internal.Counter,
				MetricName:  name,
				MetricValue: metric,
			})
		}

		err = metricsTemplate.Execute(w, allMetrics)
		if err != nil {
			internal.ErrorLog.Println("Failed to render all metrics: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
