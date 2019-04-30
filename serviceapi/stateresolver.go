package serviceapi

import "github.com/lyraproj/pcore/px"

type StateResolver interface {
	// State looks up a state that has been previously registered with the given name,
	// resolves it using the given parameters, and returns the created state object.
	State(c px.Context, name string, parameters px.OrderedMap) px.PuppetObject
}
