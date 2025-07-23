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

	srvUsecase "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecase/server"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/usecase/ping"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
)

type Server struct {
	Usecase     *srvUsecase.MetricUsecase
	PingUsecase *ping.PingUsecase
}

func NewServer(uc *srvUsecase.MetricUsecase, puc *ping.PingUsecase) *Server {
	return &Server{
		Usecase:     uc,
		PingUsecase: puc,
	}
}

func (srv *Server) GetMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "mType")
		mName := chi.URLParam(req, "mName")

		metric, err := srv.Usecase.GetMetric(req.Context(), mType, mName)
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

func (srv *Server) GetAllMetrics() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metricsToTable, err := srv.Usecase.GetAllMetrics(req.Context())
		if err != nil {
			log.Error().Err(err).Msg("failed to Get metrics")
			http.Error(res, "failed to get metrics", http.StatusNotFound)
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

		if err := srv.Usecase.UpdateMetric(req.Context(), mType, mName, val); err != nil {
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
		metric, err := srv.Usecase.GetMetric(req.Context(), jsonMetric.MType, jsonMetric.ID)
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

		if err := srv.Usecase.UpdateMetric(req.Context(), jsonMetric.MType, jsonMetric.ID, value); err != nil {
			http.Error(resp, fmt.Sprintf("invalid update metric %s: %v", jsonMetric.ID, err), http.StatusBadRequest)
			return
		}

		newMetric, err := srv.Usecase.GetMetric(req.Context(), jsonMetric.MType, jsonMetric.ID)
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

func (srv *Server) UpdatesMetricsHandlerJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var jsonMetrics serialize.MetricsList

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

		if err := srv.Usecase.UpdateMetricList(req.Context(), metrics); err != nil {
			log.Error().Err(err).Msg("failed update metrics")
			http.Error(w, fmt.Sprintf("failed update metrics: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

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
