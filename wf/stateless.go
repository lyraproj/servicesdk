package wf

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/wfapi"
)

type stateless struct {
	activity
	function interface{}
}

func NewStateless(name string, when wfapi.Condition, input, output []eval.Parameter, function interface{}) wfapi.Stateless {
	return &stateless{activity{name, when, input, output}, function}
}

func (s *stateless) Label() string {
	return `stateless ` + s.name
}

func (s *stateless) Function() interface{} {
	return s.function
}
