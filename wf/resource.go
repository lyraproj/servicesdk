package wf

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-servicesdk/wfapi"
)

type resource struct {
	activity
	state wfapi.State
}

func NewResource(name string, when wfapi.Condition, input, output []eval.Parameter, state wfapi.State) wfapi.Resource {
	return &resource{activity{name, when, input, output}, state}
}

func (r *resource) Label() string {
	return `resource ` + r.name
}

func (r *resource) State() wfapi.State {
	return r.state
}
