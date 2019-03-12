package wf

import (
	"reflect"

	"github.com/lyraproj/pcore/px"
)

type GoState struct {
	t px.ObjectType
	v reflect.Value
}

func NewGoState(t px.ObjectType, v reflect.Value) *GoState {
	return &GoState{t, v}
}

func (s *GoState) Type() px.ObjectType {
	return s.t
}

func (s *GoState) State() interface{} {
	return s.v
}

func GoStateConverter(c px.Context, state State, _ px.OrderedMap) px.PuppetObject {
	return px.WrapReflected(c, state.State().(reflect.Value)).(px.PuppetObject)
}

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
