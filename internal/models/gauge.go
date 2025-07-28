package models

type gauge struct {
	name  string
	value float64
}

func NewGauge(name string, value float64) Metric {
	return &gauge{
		name:  name,
		value: value,
	}
}

func (g *gauge) Name() string {
	return g.name
}

func (g *gauge) Type() string {
	return GaugeType
}

func (g *gauge) Value() any {
	return g.value
}

func (g *gauge) Update(mValue any) error {
	value, ok := mValue.(float64)
	if !ok {
		return ErrInvalidValueType
	}
	g.value = value
	return nil
}
