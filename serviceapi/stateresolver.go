package serviceapi

import "github.com/lyraproj/pcore/px"

type StateResolver interface {
	// State looks up a state that has been previously registered with the given name,
	// resolves it using the given input, and returns the created state object.
	State(c px.Context, name string, input px.OrderedMap) px.PuppetObject
}
