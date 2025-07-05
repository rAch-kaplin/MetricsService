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

	r.Route("/", func(r chi.Router) {
		r.Get("/", GetAllMetrics(storage))
		r.Route("/", func(r chi.Router) {
			r.Get("/value/{mType}/{mName}", GetMetric(storage))
			r.Post("/update/{mType}/{mName}/{mValue}", UpdateMetric(storage))
		})
	})

	return r
}

// FIXME other function name
func NewCounter(res http.ResponseWriter, name, value string) (mtr.Metric, error) {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		http.Error(res, "invalid value metric", http.StatusBadRequest)
		return nil, err
	}

	return mtr.NewCounter(name, val), nil
}

// FIXME other function name
func NewGauge(res http.ResponseWriter, name, value string) (mtr.Metric, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		http.Error(res, "invalid value metric", http.StatusBadRequest)
		return nil, err
	}

	return mtr.NewGauge(name, val), nil
}

func HandleUnknownMetric(res http.ResponseWriter) {
	http.Error(res, "unknown type metric!", http.StatusBadRequest)
}

func GetMetric(storage ms.Collector) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		log.Debug().Msgf("Incoming GET request: %s %s", req.Method, req.URL.Path)

		mType := chi.URLParam(req, "mType")
		mName := chi.URLParam(req, "mName")
		log.Debug().Msgf("Incoming request for metric: Type=%s, Name=%s", mType, mName)

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
			http.Error(res, "an unexpected type of metric", http.StatusInternalServerError)
			return //FIXME fix this case
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusOK)

		_, err := res.Write([]byte(valueStr))
		if err != nil {
			log.Error().Msgf("Failed to write response: %v", err)

		}

		log.Debug().Msg("the metric has been send")
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
			log.Error().Msgf("couldn't make it out HTML template: %v", err)
			http.Error(res, "Internal server error, failed html-template", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := template.Execute(res, metricsToTable); err != nil {
			log.Error().Msgf("failed complete template: %v", err)
		}

		log.Debug().Msg("the metrics has been send")
		res.WriteHeader(http.StatusOK)
	}
}

func UpdateMetric(storage ms.Collector) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		log.Debug().Msgf("Incoming POST request: %s %s", req.Method, req.URL.Path)

		mType := chi.URLParam(req, "mType")
		mName := chi.URLParam(req, "mName")
		mValue := chi.URLParam(req, "mValue")
		log.Debug().Msgf("Parsed metric: type=%s, name=%s, value=%s", mType, mName, mValue)

		if mName == "" {
			log.Error().Msgf("the metric name is not specified")
			http.Error(res, "the metric name is not specified", http.StatusBadRequest)
			return
		}

		var metric mtr.Metric
		var err error

		switch mType {
		case mtr.GaugeType:
			metric, err = NewGauge(res, mName, mValue)
		case mtr.CounterType:
			metric, err = NewCounter(res, mName, mValue)
		default:
			HandleUnknownMetric(res)
			return
		}

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err := storage.UpdateMetric(metric); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		log.Info().Msgf("Metric updated successfully: %s %s = %s", mType, mName, mValue)
		res.WriteHeader(http.StatusOK)
	}
}
