package wf

import (
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wfapi"
)

type stateHandler struct {
	activity
	api interface{}
}

func NewStateHandler(name string, when wfapi.Condition, input, output []px.Parameter, api interface{}) wfapi.StateHandler {
	return &stateHandler{activity{name, when, input, output}, api}
}

func (a *stateHandler) Label() string {
	return `stateHandler ` + a.name
}

func (a *stateHandler) Interface() interface{} {
	return a.api
}
