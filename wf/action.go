package wf

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/wfapi"
)

type action struct {
	activity
	function interface{}
}

func NewAction(name string, when wfapi.Condition, input, output []eval.Parameter, function interface{}) wfapi.Action {
	return &action{activity{name, when, input, output}, function}
}

func (s *action) Label() string {
	return `action ` + s.name
}

func (s *action) Function() interface{} {
	return s.function
}
