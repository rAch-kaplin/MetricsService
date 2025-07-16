package server

import (
	"compress/gzip"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func ConvertByType(mType, mValue string) (any, error) {
	switch mType {
	case mtr.GaugeType:
		if val, err := strconv.ParseFloat(mValue, 64); err != nil {
			return nil, fmt.Errorf("convert gauge value %s: %w", mValue, err)
		} else {
			return val, nil
		}
	case mtr.CounterType:
		if val, err := strconv.ParseInt(mValue, 10, 64); err != nil {
			return nil, fmt.Errorf("convert counter value %s: %w", mValue, err)
		} else {
			return val, nil
		}
	default:
		return nil, fmt.Errorf("unknown metric type: %s", mType)
	}
}

func GetMetric(storage col.Collector) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "mType")
		mName := chi.URLParam(req, "mName")
		value, found := storage.GetMetric(req.Context(), mType, mName)
		if !found {
			http.Error(res, fmt.Sprintf("Metric %s was not found", mName), http.StatusNotFound)
			return
		}

		var valueStr string
		switch v := value.(type) {
		case float64:
			valueStr = strconv.FormatFloat(v, 'f', -1, 64)
		case int64:
			valueStr = strconv.FormatInt(v, 10)
		default:
			http.Error(res, "an unexpected type of metric", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusOK)

		_, err := res.Write([]byte(valueStr))
		if err != nil {
			log.Error().Msgf("Failed to write response: %v", err)

		}
	}
}

func GetAllMetrics(storage col.Collector) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		allMetrics := storage.GetAllMetrics(req.Context())

		var metricsToTable []mtr.MetricTable

		for _, metric := range allMetrics {
			var valStr string
			mName := metric.Name()
			mType := metric.Type()

			switch mType {
			case mtr.GaugeType:
				val, ok := metric.Value().(float64)
				if !ok {
					log.Error().Str("metric_name", mName).Str("metric_type", mType).
						Msg("Invalid metric value type")
					continue
				}
				valStr = strconv.FormatFloat(val, 'f', -1, 64)

			case mtr.CounterType:
				val, ok := metric.Value().(int64)
				if !ok {
					log.Error().Str("metric_name", mName).Str("metric_type", mType).
						Msg("Invalid metric value type")
					continue
				}
				valStr = strconv.FormatInt(val, 10)
			}

			metricsToTable = append(metricsToTable, mtr.MetricTable{
				Name:  mName,
				Type:  mType,
				Value: valStr,
			})
		}

		const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Metric</h1>
    <table border="1">
        <thead>
            <tr>
                <th>Name of Metric</th>
                <th>Type</th>
                <th>Value</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Type}}</td>
                <td>{{.Value}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>
`

		template, err := template.New("Metrics").Parse(htmlTemplate)
		if err != nil {
			log.Error().Msgf("couldn't make it out HTML template: %v", err)
			http.Error(res, "Internal server error, failed html-template", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := template.Execute(res, metricsToTable); err != nil {
			log.Error().Msgf("failed complete template: %v", err)
		}
	}
}

func UpdateMetric(storage col.Collector) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "mType")
		mName := chi.URLParam(req, "mName")
		mValue := chi.URLParam(req, "mValue")

		if mName == "" {
			log.Error().
				Str("metric_type", mType).
				Msg("Metric name is not specified")

			http.Error(res, "the metric name is not specified", http.StatusBadRequest)
			return
		}

		val, err := ConvertByType(mType, mValue)
		if err != nil {
			log.Error().
				Err(err).
				Str("metric_type", mType).
				Str("metric_value", mValue).
				Msg("Failed to convert metric value")

			http.Error(res, fmt.Sprintf("invalid metric value: %v", err), http.StatusBadRequest)
			return
		}

		if err := storage.UpdateMetric(req.Context(), mType, mName, val); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		log.Info().
			Str("metric_type", mType).
			Str("metric_name", mName).
			Interface("value", val).
			Msg("Metric updated successfully")
	}
}

func FillMetricValueFromStorage(ctx context.Context, storage col.Collector, metric *mtr.Metrics) bool {
	value, ok := storage.GetMetric(ctx, metric.MType, metric.ID)
	if !ok {
		return false
	}

	switch v := value.(type) {
	case float64:
		metric.Value = &v
	case int64:
		metric.Delta = &v
	default:
		log.Error().
			Str("metricType", metric.MType).
			Str("metricName", metric.ID).
			Msg("unsupported metric type")

		return false
	}

	return true
}

func GetMetricsHandlerJSON(storage col.Collector) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		var metric mtr.Metrics

		log.Info().Msg("GetMetricsHandlerJSON called")
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(resp, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		if err := easyjson.UnmarshalFromReader(req.Body, &metric); err != nil {
			http.Error(resp, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
			return
		}

		if !FillMetricValueFromStorage(req.Context(), storage, &metric) {
			http.Error(resp, fmt.Sprintf("metric %s not found", metric.ID), http.StatusNotFound)
			return
		}
		log.Info().Msgf("Received metric request: %+v", metric)

		resp.Header().Set("Content-Type", "application/json")
		if _, err := easyjson.MarshalToWriter(&metric, resp); err != nil {
			http.Error(resp, fmt.Sprintf("failed to encode json: %v", err), http.StatusInternalServerError)
			return
		}
		log.Info().Msg("Metric successfully returned\n\n")
	}
}

func UpdateMetricsHandlerJSON(storage col.Collector) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		var reader io.Reader
		if req.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(req.Body)
			if err != nil {
				http.Error(resp, "failed to create gzip reader", http.StatusBadRequest)
				return
			}
			defer func() {
				if err := gz.Close(); err != nil {
					log.Error().Err(err).Msg("failed close gz reader")
				}
			}()
			reader = gz
		} else {
			reader = req.Body
		}

		var metric mtr.Metrics

		log.Info().Msg("UpdateMetricsHandlerJSON called")
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(resp, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		if err := easyjson.UnmarshalFromReader(reader, &metric); err != nil {
			http.Error(resp, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
			return
		}

		var value any
		if metric.MType == mtr.GaugeType {
			value = *metric.Value
		} else {
			value = *metric.Delta
		}

		if err := storage.UpdateMetric(req.Context(), metric.MType, metric.ID, value); err != nil {
			http.Error(resp, fmt.Sprintf("invalid update metric %s: %v", metric.ID, err), http.StatusBadRequest)
		}

		if !FillMetricValueFromStorage(req.Context(), storage, &metric) {
			http.Error(resp, fmt.Sprintf("metric %s not found", metric.ID), http.StatusNotFound)
			return
		}

		resp.Header().Set("Content-Type", "application/json")
		if _, err := easyjson.MarshalToWriter(&metric, resp); err != nil {
			http.Error(resp, fmt.Sprintf("failed to encode json: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func UpdatesMetricsHandlerJSON(storage col.Collector) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var jsonMetrics mtr.MetricsList

		log.Info().Msg("UpdatesMetricsHandlerJSON called")
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		if err := easyjson.UnmarshalFromReader(req.Body, &jsonMetrics); err != nil {
			http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
			return
		}

		metrics, err := converter.ConvertMetrics(jsonMetrics)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid convert metrics: %v", err), http.StatusBadRequest)
			return
		}

		for _, metric := range metrics {
			if err := storage.UpdateMetric(req.Context(), metric.Type(), metric.Name(), metric.Value()); err != nil {
				http.Error(w, fmt.Sprintf("failed update metric %v", err), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func PingHandler(col col.Collector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := col.Ping(r.Context()); err != nil {
			http.Error(w, "can't ping DB", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
