package serviceapi

import "github.com/lyraproj/puppet-evaluator/eval"

type Invokable interface {
	// Invoke will call a method with the given name on the object identified by the given
	// identifier and return the result.
	Invoke(c eval.Context, identifier, name string, arguments ...eval.Value) eval.Value
}
