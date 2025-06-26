package metrics

type counter struct {
	name  string
	value int64
}

func NewCounter(name string) Metric {
	return &counter{
		name: name,
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
