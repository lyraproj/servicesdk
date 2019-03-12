package wf

import (
	"github.com/lyraproj/pcore/px"
)

type Action interface {
	Activity

	Function() interface{}
}

type action struct {
	activity
	function interface{}
}

func MakeAction(name string, when Condition, input, output []px.Parameter, function interface{}) Action {
	return &action{activity{name, when, input, output}, function}
}

func (s *action) Label() string {
	return `action ` + s.name
}

func (s *action) Function() interface{} {
	return s.function
}
