package converter

import (
	"fmt"
	"strconv"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	pb "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/grpc-metrics"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
	"github.com/rs/zerolog/log"
)

func ConvertByType(mType, mValue string) (any, error) {
	switch mType {
	case models.GaugeType:
		if val, err := strconv.ParseFloat(mValue, 64); err != nil {
			return nil, fmt.Errorf("convert gauge value %s: %w", mValue, err)
		} else {
			return val, nil
		}
	case models.CounterType:
		if val, err := strconv.ParseInt(mValue, 10, 64); err != nil {
			return nil, fmt.Errorf("convert counter value %s: %w", mValue, err)
		} else {
			return val, nil
		}
	default:
		return nil, fmt.Errorf("unknown metric type: %s", mType)
	}
}

func ConvertToMetricTable(src []models.Metric) ([]models.MetricTable, error) {
	converted := make([]models.MetricTable, 0, len(src))

	for _, metric := range src {
		var valStr string
		mName := metric.Name()
		mType := metric.Type()

		switch mType {
		case models.GaugeType:
			val, ok := metric.Value().(float64)
			if !ok {
				log.Error().Str("metric_name", mName).Str("metric_type", mType).
					Msg("Invalid metric value type")

				return nil, fmt.Errorf("invalid metric value type: %s", mType)
			}
			valStr = strconv.FormatFloat(val, 'f', -1, 64)

		case models.CounterType:
			val, ok := metric.Value().(int64)
			if !ok {
				log.Error().Str("metric_name", mName).Str("metric_type", mType).
					Msg("Invalid metric value type")

				return nil, fmt.Errorf("invalid metric value type: %s", mType)
			}
			valStr = strconv.FormatInt(val, 10)
		}

		converted = append(converted, models.MetricTable{
			Name:  mName,
			Type:  mType,
			Value: valStr,
		})
	}

	return converted, nil
}

func ConvertMetrics(src serialize.MetricsList) ([]models.Metric, error) {
	converted := make([]models.Metric, 0, len(src))

	for _, m := range src {
		metric, err := сonvertMetric(m)
		if err != nil {
			log.Error().Err(err).Msg("failed convert metric")
			return nil, fmt.Errorf("metric failed convert %+v", m)
		}

		converted = append(converted, metric)
	}

	return converted, nil
}

func сonvertMetric(src serialize.Metric) (models.Metric, error) {
	var converted models.Metric

	switch src.MType {
	case models.GaugeType:
		if src.Value == nil {
			return nil, fmt.Errorf("nil gauge value for ID: %s", src.ID)
		}
		converted = models.NewGauge(src.ID, *src.Value)

	case models.CounterType:
		if src.Delta == nil {
			return nil, fmt.Errorf("nil counter value for ID: %s", src.ID)
		}
		converted = models.NewCounter(src.ID, *src.Delta)
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", src.MType)
	}

	return converted, nil
}

func сonvertToSerialization(src models.Metric) (serialize.Metric, error) {
	var converted serialize.Metric

	switch src.Type() {
	case models.GaugeType:
		value, ok := src.Value().(float64)
		if !ok {
			return serialize.Metric{}, fmt.Errorf("invalid gauge value: %v", src.Value())
		}
		converted = serialize.Metric{
			ID:    src.Name(),
			MType: src.Type(),
			Value: &value,
		}

	case models.CounterType:
		delta, ok := src.Value().(int64)
		if !ok {
			return serialize.Metric{}, fmt.Errorf("invalid counter value: %v", src.Value())
		}
		converted = serialize.Metric{
			ID:    src.Name(),
			MType: src.Type(),
			Delta: &delta,
		}

	default:
		return serialize.Metric{}, fmt.Errorf("unknown metric type: %s", src.Type())
	}

	return converted, nil
}

func ConvertToSerialization(src []models.Metric) ([]serialize.Metric, error) {
	converted := make([]serialize.Metric, 0, len(src))

	for _, mtr := range src {
		metric, err := сonvertToSerialization(mtr)
		if err != nil {
			return nil, fmt.Errorf("metric failed convert %+v", mtr)
		}

		converted = append(converted, metric)
	}

	return converted, nil
}

func convertToProtoMetric(src models.Metric) *pb.Metric {
	switch src.Type() {
	case models.GaugeType:
		value, ok := src.Value().(float64)
		if !ok {
			return nil
		}
		return &pb.Metric{
			Id:       src.Name(),
			MType:    src.Type(),
			MetricValue: &pb.Metric_Value{
				Value: value,
			},
		}

	case models.CounterType:
		delta, ok := src.Value().(int64)
		if !ok {
			return nil
		}
		return &pb.Metric{
			Id:       src.Name(),
			MType:    src.Type(),
			MetricValue: &pb.Metric_Delta{
				Delta: delta,
			},
		}
	}

	return nil
}

func ConvertToProtoMetrics(src []models.Metric) ([]*pb.Metric, error) {
	converted := make([]*pb.Metric, 0, len(src))

	for _, m := range src {
		metric := convertToProtoMetric(m)
		if metric == nil {
			return nil, fmt.Errorf("failed to convert metric %+v", m)
		}
		converted = append(converted, metric)
	}

	return converted, nil
}
