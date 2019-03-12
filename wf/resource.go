package wf

import (
	"github.com/lyraproj/pcore/px"
)

type State interface {
	Type() px.ObjectType
	State() interface{}
}

type StateConverter func(ctx px.Context, state State, input px.OrderedMap) px.PuppetObject

type Resource interface {
	Activity

	ExternalId() string

	State() State
}

type resource struct {
	activity
	state State
	extId string
}

func MakeResource(name string, when Condition, input, output []px.Parameter, extId string, state State) Resource {
	return &resource{activity{name, when, input, output}, state, extId}
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
