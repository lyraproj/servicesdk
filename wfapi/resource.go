package wfapi

import "github.com/lyraproj/pcore/px"

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
