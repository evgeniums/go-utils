package api_server

// Generic parameter of requests and responses.
type Parameter interface {
	Name() string
	Value() interface{}
	SetValue(val interface{})
}

// Base interface for types with parameters.
type WithParameters interface {
	AddParameter(param Parameter)
	GetParameter(name string) Parameter
	SetParameter(name string, value interface{})
	HasParameter(name string) bool
	ParameterValue(name string) interface{}
}

// Base parameter type.
type ParameterBase struct {
	name  string
	value interface{}
}

func NewParameter(name string, value interface{}) Parameter {
	return &ParameterBase{name: name, value: value}
}

func (p *ParameterBase) Name() string {
	return p.name
}

func (p *ParameterBase) Value() interface{} {
	return p.value
}

func (p *ParameterBase) SetValue(val interface{}) {
	p.value = val
}
