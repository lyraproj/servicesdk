package serviceapi

import "github.com/puppetlabs/go-evaluator/eval"

type StateResolver interface {
	// State looks up a state that has been previously registered with the given name,
	// resolves it using the given input, and returns the created state object.
	State(c eval.Context, name string, input eval.OrderedMap) eval.PuppetObject
}
