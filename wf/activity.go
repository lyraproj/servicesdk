package wf

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-servicesdk/wfapi"
)

type activity struct {
	name   string
	when   wfapi.Condition
	input  []eval.Parameter
	output []eval.Parameter
}

func (a *activity) When() wfapi.Condition {
	return a.when
}

func (a *activity) Name() string {
	return a.name
}

func (a *activity) Input() []eval.Parameter {
	return a.input
}

func (a *activity) Output() []eval.Parameter {
	return a.output
}
