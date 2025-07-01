package server

import (
	"net/http"
	"strconv"
	"strings"

	ms "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/MemStorage"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/logger"
)

func HandleCounter(res http.ResponseWriter, name, value string) (mtr.Metric, error) {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		http.Error(res, "invalid value metric", http.StatusBadRequest)
		return nil, err
	}

	return mtr.NewCounter(name, val), nil
}

func HandleGauge(res http.ResponseWriter, name, value string) (mtr.Metric, error) {
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

func removeEmptyStrings(url []string) []string {
	result := make([]string, 0, len(url))
	for _, str := range url {
		if str != "" {
			result = append(result, str)
		}
	}

	return result
}

func MainHandle(storage ms.Collector) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		log.Debug("Incoming request: %s %s", req.Method, req.URL.Path)

		if req.Method != http.MethodPost {
			http.Error(res, "use only POST request", http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get("Content-Type") != "text/plain" {
			http.Error(res, "use only text/plain", http.StatusBadRequest)
			return
		}

		parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
		parts = removeEmptyStrings(parts)

		if len(parts) != 4 || parts[0] != "update" {
			http.Error(res, "invalid request", http.StatusNotFound)
			return
		}

		metricType, name, value := parts[1], parts[2], parts[3]
		log.Debug("Parsed metric: type=%s, name=%s, value=%s", metricType, name, value)

		if name == "" {
			http.Error(res, "metric name is missing", http.StatusNotFound)
			return
		}

		var metric mtr.Metric
		var err error

		switch metricType {
		case mtr.GaugeType:
			metric, err = HandleGauge(res, name, value)
		case mtr.CounterType:
			metric, err = HandleCounter(res, name, value)
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

		log.Info("Metric updated successfully: %s %s = %s", metricType, name, value)
		res.WriteHeader(http.StatusOK)
	}
}
