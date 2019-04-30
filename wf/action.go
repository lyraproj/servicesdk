package wf

import (
	"github.com/lyraproj/pcore/px"
)

type Action interface {
	Step

	Function() interface{}
}

type action struct {
	step
	function interface{}
}

func MakeAction(name string, when Condition, parameters, returns []px.Parameter, function interface{}) Action {
	return &action{step{name, when, parameters, returns}, function}
}

func (s *action) Label() string {
	return `action ` + s.name
}

func (s *action) Function() interface{} {
	return s.function
}
