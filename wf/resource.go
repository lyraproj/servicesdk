package wf

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/wfapi"
)

type resource struct {
	activity
	state wfapi.State
	extId string
}

func NewResource(name string, when wfapi.Condition, input, output []eval.Parameter, extId string, state wfapi.State) wfapi.Resource {
	return &resource{activity{name, when, input, output}, state, extId}
}

func (r *resource) ExternalId() string {
	return r.extId
}

func (r *resource) Label() string {
	return `resource ` + r.name
}

func (r *resource) State() wfapi.State {
	return r.state
}
