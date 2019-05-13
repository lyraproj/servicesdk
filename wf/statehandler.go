package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

type StateHandler interface {
	Step

	Interface() interface{}
}

type stateHandler struct {
	step
	api interface{}
}

func MakeStateHandler(name string, origin issue.Location, when Condition, parameters, returns []px.Parameter, api interface{}) StateHandler {
	return &stateHandler{step{name, origin, when, parameters, returns}, api}
}

func (a *stateHandler) Label() string {
	return `stateHandler ` + a.name
}

func (a *stateHandler) Interface() interface{} {
	return a.api
}
