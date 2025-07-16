package converter

import (
	"fmt"

	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
)

func ConvertMetrics(src []mtr.Metrics) ([]mtr.Metric, error) {
	converted := make([]mtr.Metric, 0, len(src))

	for _, m := range src {
		switch m.MType {
		case mtr.GaugeType:
			if m.Value == nil {
				return nil, fmt.Errorf("nil gauge value for ID: %s", m.ID)
			}
			converted = append(converted, mtr.NewGauge(m.ID, *m.Value))

		case mtr.CounterType:
			if m.Delta == nil {
				return nil, fmt.Errorf("nil counter value for ID: %s", m.ID)
			}
			converted = append(converted, mtr.NewCounter(m.ID, *m.Delta))
		default:
			return nil, fmt.Errorf("unsupported metric type: %s", m.MType)
		}
	}

	return converted, nil
}

func ConvertMetric(src mtr.Metrics) (mtr.Metric, error) {
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
