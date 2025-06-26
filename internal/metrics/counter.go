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

func (c *counter) Value() interface{} {
	return c.value
}

func (c *counter) SetValue(v interface{}) error {
	val, ok := v.(int64)
	if !ok {
		return ErrInvalidValueType
	}
	c.value = val
	return nil
}
