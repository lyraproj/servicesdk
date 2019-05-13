package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

type State interface {
	Type() px.ObjectType
	State() interface{}
}

type StateConverter func(ctx px.Context, state State, parameters px.OrderedMap) px.PuppetObject

type Resource interface {
	Step

	ExternalId() string

	State() State
}

type resource struct {
	step
	state State
	extId string
}

func MakeResource(name string, origin issue.Location, when Condition, parameters, returns []px.Parameter, extId string, state State) Resource {
	return &resource{step{name, origin, when, parameters, returns}, state, extId}
}

func (r *resource) ExternalId() string {
	return r.extId
}

func (r *resource) Label() string {
	return `resource ` + r.name
}

func (r *resource) State() State {
	return r.state
}
