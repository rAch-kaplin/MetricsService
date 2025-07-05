package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/mem-storage"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

type MetricTable struct {
	Name  string
	Type  string
	Value string
}

func NewRouter(storage ms.Collector) http.Handler {
	r := chi.NewRouter()

	r.Use(WithLogging)

	r.Route("/", func(r chi.Router) {
		r.Get("/", GetAllMetrics(storage))
		r.Route("/", func(r chi.Router) {
			r.Get("/value/{mType}/{mName}", GetMetric(storage))
			r.Post("/update/{mType}/{mName}/{mValue}", UpdateMetric(storage))
		})
	})

	return r
}

func newCounter(res http.ResponseWriter, name, value string) (mtr.Metric, error) {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Error().Err(err).Str("value", value).Msg("invalid counter metric value")
		http.Error(res, "invalid value metric", http.StatusBadRequest)
		return nil, err
	}

	return mtr.NewCounter(name, val), nil
}

func newGauge(res http.ResponseWriter, name, value string) (mtr.Metric, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Error().Err(err).Str("value", value).Msg("invalid gauge metric value")
		http.Error(res, "invalid value metric", http.StatusBadRequest)
		return nil, err
	}

	return mtr.NewGauge(name, val), nil
}

func HandleUnknownMetric(res http.ResponseWriter, mType string) {
	log.Error().Str("metricType", mType).Msg("unknown metric type")
	http.Error(res, "unknown type metric!", http.StatusBadRequest)
}

func GetMetric(storage ms.Collector) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "mType")
		mName := chi.URLParam(req, "mName")
		value, found := storage.GetMetric(mType, mName)
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
			log.Error().Msg("unexpected type of metric")
			http.Error(res, "unexpected type of metric", http.StatusInternalServerError)
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusOK)

		_, err := res.Write([]byte(valueStr))
		if err != nil {
			log.Error().Msgf("Failed to write response: %v", err)

		}
	}
}

func GetAllMetrics(storage ms.Collector) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		gauges, counters := storage.GetAllMetrics()

		var metricsToTable []MetricTable

		for name, value := range gauges {
			metricsToTable = append(metricsToTable, MetricTable{
				Name:  name,
				Type:  mtr.GaugeType,
				Value: strconv.FormatFloat(value, 'f', -1, 64),
			})
		}
		for name, value := range counters {
			metricsToTable = append(metricsToTable, MetricTable{
				Name:  name,
				Type:  mtr.CounterType,
				Value: strconv.FormatInt(value, 10),
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
			log.Error().Err(err).Msg("failed to parse HTML template")
			http.Error(res, "Internal server error, failed html-template", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := template.Execute(res, metricsToTable); err != nil {
			log.Error().Err(err).Msg("failed to execute HTML template")
		}

		res.WriteHeader(http.StatusOK)
	}
}

func UpdateMetric(storage ms.Collector) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "mType")
		mName := chi.URLParam(req, "mName")
		mValue := chi.URLParam(req, "mValue")

		if mName == "" {
			log.Error().Msgf("the metric name is not specified")
			http.Error(res, "the metric name is not specified", http.StatusBadRequest)
			return
		}

		var metric mtr.Metric
		var err error

		switch mType {
		case mtr.GaugeType:
			metric, err = newGauge(res, mName, mValue)
		case mtr.CounterType:
			metric, err = newCounter(res, mName, mValue)
		default:
			HandleUnknownMetric(res, mType)
			return
		}

		if err != nil {
			return
		}

		if err := storage.UpdateMetric(metric); err != nil {
			log.Error().Err(err).Msg("failed to update metric in storage")
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}
