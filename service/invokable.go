package service

import "github.com/puppetlabs/go-evaluator/eval"

type Invokable interface {
	// Invoke will call a method with the given name on the object identified by the given
	// identifier and return the result.
	Invoke(ctx eval.Context, identifier eval.TypedName, name string, arguments eval.OrderedMap) eval.Value
}
