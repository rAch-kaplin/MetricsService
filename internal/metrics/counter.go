package metrics

type counter struct {
	name  string
	value int64
}

func NewCounter(name string, value int64) Metric {
	return &counter{
		name:  name,
		value: value,
	}
}

func (c *counter) Name() string {
	return c.name
}

func (c *counter) Type() string {
	return CounterType
}

func (c *counter) Value() any {
	return c.value
}

func (c *counter) Update(mType, mName string, mValue any) error {
	if mType != c.Type() {
		return ErrInvalidMetricsType
	}

	mtrValue, ok := mValue.(int64)
	if !ok {
		return ErrInvalidValueType
	}

	c.value += mtrValue

	return nil
}

func (c *counter) SetValue(v any) error {
	val, ok := v.(int64)
	if !ok {
		return ErrInvalidValueType
	}
	c.value = val
	return nil
}
