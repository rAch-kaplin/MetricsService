package converter

import (
	"fmt"
	"strconv"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
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

func ConvertMetrics(src serialize.MetricsList) ([]models.Metric, error) {
	converted := make([]models.Metric, 0, len(src))

	for _, m := range src {
		metric, err := ConvertMetric(m)
		if err != nil {
			log.Error().Err(err).Msg("failed convert metric")
			return nil, fmt.Errorf("metric failed convert %+v", m)
		}

		converted = append(converted, metric)
	}

	return converted, nil
}

func ConvertMetric(src serialize.Metric) (models.Metric, error) {
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
