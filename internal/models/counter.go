package models

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

func (c *counter) Update(mValue any) error {
	value, ok := mValue.(int64)
	if !ok {
		return ErrInvalidValueType
	}
	c.value += value
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
