package wf

import (
	"github.com/lyraproj/pcore/px"
)

type StateHandler interface {
	Activity

	Interface() interface{}
}

type stateHandler struct {
	activity
	api interface{}
}

func MakeStateHandler(name string, when Condition, input, output []px.Parameter, api interface{}) StateHandler {
	return &stateHandler{activity{name, when, input, output}, api}
}

func (a *stateHandler) Label() string {
	return `stateHandler ` + a.name
}

func (a *stateHandler) Interface() interface{} {
	return a.api
}
