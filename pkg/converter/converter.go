package converter

import (
	"fmt"

	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
	"github.com/rs/zerolog/log"
)

func ConvertMetrics(src serialize.MetricsList) ([]mtr.Metric, error) {
	converted := make([]mtr.Metric, 0, len(src))

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

func ConvertMetric(src serialize.Metric) (mtr.Metric, error) {
	var converted mtr.Metric

	switch src.MType {
	case mtr.GaugeType:
		if src.Value == nil {
			return nil, fmt.Errorf("nil gauge value for ID: %s", src.ID)
		}
		converted = mtr.NewGauge(src.ID, *src.Value)

	case mtr.CounterType:
		if src.Delta == nil {
			return nil, fmt.Errorf("nil counter value for ID: %s", src.ID)
		}
		converted = mtr.NewCounter(src.ID, *src.Delta)
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", src.MType)
	}

	return converted, nil
}
