// @Title MetricsService API handlers
// @Description This file contains the server handlers for the MetricsService API.
// @Author rAch-kaplin
// @Version 1.0.0
// @Since 2025-07-29

package server

import (
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/ping"
	srvUsecase "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecases/server"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
)

type Server struct {
	MetricUsecase *srvUsecase.MetricUsecase
	PingUsecase   *ping.PingUsecase
}

func NewServer(uc *srvUsecase.MetricUsecase, puc *ping.PingUsecase) *Server {
	return &Server{
		MetricUsecase: uc,
		PingUsecase:   puc,
	}
}

// @Title GetMetric
// @Description Get a metric by type and name from URL parameters
// @Tags metrics
// @Produces text/plain
// @Param mType path string true "Metric type"
// @Param mName path string true "Metric name"
// @Success 200 {string} string "Metric value"
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "Metric not found"
// @Failure 500 {string} string "Internal server error"
// @Router /value/{mType}/{mName} [GET]
func (srv *Server) GetMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "mType")
		mName := chi.URLParam(req, "mName")

		metric, err := srv.MetricUsecase.GetMetric(req.Context(), mType, mName)
		if err != nil {
			log.Error().Err(err).Msg("can't get valid metric")
			http.Error(res, fmt.Sprintf("Metric %s was not found", mName), http.StatusNotFound)
			return
		}

		var valueStr string
		switch v := metric.Value().(type) {
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

		_, err = res.Write([]byte(valueStr))
		if err != nil {
			log.Error().Msgf("Failed to write response: %v", err)

		}
	}
}

// @Title GetAllMetrics
// @Description Get all metrics
// @Tags metrics
// @Produces text/html
// @Success 200 {string} string "Metrics table"
// @Failure 404 {string} string "Metrics not found"
// @Failure 500 {string} string "Internal server error"
// @Router / [GET]
func (srv *Server) GetAllMetrics() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metrics, err := srv.MetricUsecase.GetAllMetrics(req.Context())
		if err != nil {
			log.Error().Err(err).Msg("failed to Get metrics")
			http.Error(res, "failed to get metrics", http.StatusNotFound)
			return
		}

		metricsToTable, err := converter.ConvertToMetricTable(metrics)
		if err != nil {
			log.Error().Err(err).Msg("failed to convert metrics to table")
			http.Error(res, "failed to convert metrics to table", http.StatusInternalServerError)
			return
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

// @Title UpdateMetric
// @Description Update a metric by type, name and value from URL parameters
// @Tags metrics
// @Produces text/plain
// @Param mType path string true "Metric type"
// @Param mName path string true "Metric name"
// @Param mValue path string true "Metric value"
// @Success 200 {string} string "Metric updated successfully"
// @Failure 400 {string} string "Bad request - invalid metric value or name not specified"
// @Failure 500 {string} string "Internal server error"
// @Router /update/{mType}/{mName}/{mValue} [POST]
func (srv *Server) UpdateMetric() http.HandlerFunc {
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

		val, err := converter.ConvertByType(mType, mValue)
		if err != nil {
			log.Error().
				Err(err).
				Str("metric_type", mType).
				Str("metric_value", mValue).
				Msg("Failed to convert metric value")

			http.Error(res, fmt.Sprintf("invalid metric value: %v", err), http.StatusBadRequest)
			return
		}

		if err := srv.MetricUsecase.UpdateMetric(req.Context(), mType, mName, val); err != nil {
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

// @Title GetMetricsHandlerJSON
// @Description Get a metric by type and name
// @Tags metrics
// @Produces application/json
// @Accept application/json
// @Failure 400 {string} string "Invalid JSON body"
// @Failure 415 {string} string "Unsupported media type"
// @Failure 404 {string} string "Metric not found"
// @Failure 500 {string} string "Internal server error"
// @Router /value [POST]
func (srv *Server) GetMetricsHandlerJSON() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		var jsonMetric serialize.Metric

		log.Info().Msg("GetMetricsHandlerJSON called")
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(resp, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		if err := easyjson.UnmarshalFromReader(req.Body, &jsonMetric); err != nil {
			http.Error(resp, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
			return
		}

		log.Debug().Str("type", jsonMetric.MType).Str("name", jsonMetric.ID).Msg("")
		metric, err := srv.MetricUsecase.GetMetric(req.Context(), jsonMetric.MType, jsonMetric.ID)
		if err != nil {
			log.Error().Err(err).Msg("can't get valid metric")
			http.Error(resp, "can't get valid metric", http.StatusNotFound)
			return
		}

		if err := jsonMetric.SetValue(metric.Value()); err != nil {
			log.Error().Err(err).Msg("can't set new value")
			http.Error(resp, "can't set new value", http.StatusInternalServerError)
			return
		}

		resp.Header().Set("Content-Type", "application/json")
		if _, err := easyjson.MarshalToWriter(&jsonMetric, resp); err != nil {
			http.Error(resp, fmt.Sprintf("failed to encode json: %v", err), http.StatusInternalServerError)
			return
		}
		log.Info().Msg("Metric successfully returned\n\n")
	}
}

// @Title UpdateMetricsHandlerJSON
// @Description Update a metric by type and name
// @Tags metrics
// @Produces application/json
// @Accept application/json
// @Failure 400 {string} string "Invalid JSON body"
// @Failure 415 {string} string "Unsupported media type"
// @Failure 404 {string} string "Metric not found"
// @Failure 500 {string} string "Internal server error"
// @Router /update [POST]
func (srv *Server) UpdateMetricsHandlerJSON() http.HandlerFunc {
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

		var jsonMetric serialize.Metric

		log.Info().Msg("UpdateMetricsHandlerJSON called")
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(resp, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		if err := easyjson.UnmarshalFromReader(reader, &jsonMetric); err != nil {
			http.Error(resp, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
			return
		}

		value, err := jsonMetric.GetValue()
		if err != nil {
			log.Error().Err(err).Msg("can't get value")
			http.Error(resp, "can't get value", http.StatusBadRequest)
			return
		}

		if err := srv.MetricUsecase.UpdateMetric(req.Context(), jsonMetric.MType, jsonMetric.ID, value); err != nil {
			http.Error(resp, fmt.Sprintf("invalid update metric %s: %v", jsonMetric.ID, err), http.StatusBadRequest)
			return
		}

		newMetric, err := srv.MetricUsecase.GetMetric(req.Context(), jsonMetric.MType, jsonMetric.ID)
		if err != nil {
			log.Error().Err(err).Msg("can't get new value metric")
			http.Error(resp, "can't get new value metric", http.StatusNotFound)
			return
		}

		err = jsonMetric.SetValue(newMetric.Value())
		if err != nil {
			http.Error(resp, fmt.Sprintf("metric %s not found", jsonMetric.ID), http.StatusNotFound)
			return
		}

		resp.Header().Set("Content-Type", "application/json")
		if _, err := easyjson.MarshalToWriter(&jsonMetric, resp); err != nil {
			http.Error(resp, fmt.Sprintf("failed to encode json: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// @Title UpdatesMetricsHandlerJSON
// @Description Update a list of metrics
// @Tags metrics
// @Produces application/json
// @Accept application/json
// @Success 200 {string} string "Metrics updated successfully"
// @Failure 400 {string} string "Invalid JSON body"
// @Failure 415 {string} string "Unsupported media type"
// @Failure 500 {string} string "Internal server error"
// @Router /updates [POST]
func (srv *Server) UpdatesMetricsHandlerJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var reader io.Reader
		if req.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(req.Body)
			if err != nil {
				http.Error(w, "failed to create gzip reader", http.StatusBadRequest)
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

		var jsonMetrics serialize.MetricsList

		log.Info().Msg("UpdatesMetricsHandlerJSON called")
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		if err := easyjson.UnmarshalFromReader(reader, &jsonMetrics); err != nil {
			http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
			return
		}

		metrics, err := converter.ConvertMetrics(jsonMetrics)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid convert metrics: %v", err), http.StatusBadRequest)
			return
		}

		if err := srv.MetricUsecase.UpdateMetricList(req.Context(), metrics); err != nil {
			log.Error().Err(err).Msg("failed update metrics")
			http.Error(w, fmt.Sprintf("failed update metrics: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Title PingHandler
// @Description Check if the database is reachable
// @Tags metrics
// @Produces text/plain
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Database connection failed"
// @Router /ping [GET]
func (srv *Server) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := srv.PingUsecase.Check(r.Context()); err != nil {
			log.Error().Err(err).Msg("failed ping")
			http.Error(w, "can't ping DB", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
