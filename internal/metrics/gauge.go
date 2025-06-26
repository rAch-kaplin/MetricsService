package metrics

type gauge struct {
	name  string
	value float64
}
func NewGauge(name string) Metric {
	return &gauge{
		name: name,
	}
}

func (g *gauge) Name() string {
	return g.name
}

func (g *gauge) Type() string {
	return GaugeType
}

func (g *gauge) Value() interface{} {
	return g.value
}
