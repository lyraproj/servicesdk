package serviceapi

import "github.com/lyraproj/pcore/px"

type Parameter interface {
	px.Value

	// Name of the parameter
	Name() string

	// Alias to use inside the step that uses this parameter or the empty
	// string if no alias exists.
	Alias() string

	// The Type of the parameter.
	Type() px.Type

	// The parameter value, or nil if parameter has no value. An Undef is
	// considered a valid value.
	Value() px.Value
}

// NewParameter creates a new parameter instance
var NewParameter func(name, alias string, typ px.Type, value px.Value) Parameter
