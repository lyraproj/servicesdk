package wfapi

import "github.com/lyraproj/puppet-evaluator/eval"

type State interface {
	Type() eval.ObjectType
	State() interface{}
}

type StateConverter func(ctx eval.Context, state State, input eval.OrderedMap) eval.PuppetObject

type Resource interface {
	Activity

	ExternalId() string

	State() State
}
