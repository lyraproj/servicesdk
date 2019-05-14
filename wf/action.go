package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/servicesdk/serviceapi"
)

type Action interface {
	Step

	Function() interface{}
}

type action struct {
	step
	function interface{}
}

func MakeAction(name string, origin issue.Location, when Condition, parameters, returns []serviceapi.Parameter, function interface{}) Action {
	return &action{step{name, origin, when, parameters, returns}, function}
}

func (s *action) Label() string {
	return `action ` + s.name
}

func (s *action) Function() interface{} {
	return s.function
}
