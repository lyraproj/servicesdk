package serviceapi

import "github.com/lyraproj/pcore/px"

type Invokable interface {
	// Invoke will call a method with the given name on the object identified by the given
	// identifier and return the result.
	Invoke(c px.Context, identifier, name string, arguments ...px.Value) px.Value
}
