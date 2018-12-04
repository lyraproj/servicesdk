package wf

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/wfapi"
)

type action struct {
	activity
	api interface{}
}

func NewAction(name string, when wfapi.Condition, input, output []eval.Parameter, api interface{}) wfapi.Action {
	return &action{activity{name, when, input, output}, api}
}

func (a *action) Label() string {
	return `action ` + a.name
}

func (a *action) Interface() interface{} {
	return a.api
}
