package wf

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-servicesdk/wfapi"
)

type stateless struct {
	activity
	doer wfapi.Doer
}

func NewStateless(name string, when wfapi.Condition, input, output []eval.Parameter, doer wfapi.Doer) wfapi.Stateless {
	return &stateless{activity{name, when, input, output}, doer}
}

func (s *stateless) Label() string {
	return `stateless ` + s.name
}

func (s *stateless) Interface() wfapi.Doer {
	return s.doer
}
