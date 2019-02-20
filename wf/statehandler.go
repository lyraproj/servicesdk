package wf

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/wfapi"
)

type stateHandler struct {
	activity
	api interface{}
}

func NewStateHandler(name string, when wfapi.Condition, input, output []eval.Parameter, api interface{}) wfapi.StateHandler {
	return &stateHandler{activity{name, when, input, output}, api}
}

func (a *stateHandler) Label() string {
	return `stateHandler ` + a.name
}

func (a *stateHandler) Interface() interface{} {
	return a.api
}
