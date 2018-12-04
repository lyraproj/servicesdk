package wf

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/wfapi"
)

type iterator struct {
	activity
	style     wfapi.IterationStyle
	producer  wfapi.Activity
	over      []eval.Parameter
	variables []eval.Parameter
}

func NewIterator(name string, when wfapi.Condition, input, output []eval.Parameter, style wfapi.IterationStyle, producer wfapi.Activity, over []eval.Parameter, variables []eval.Parameter) wfapi.Iterator {
	return &iterator{activity{name, when, input, output}, style, producer, over, variables}
}

func (it *iterator) Label() string {
	return `iterator ` + it.name
}

func (it *iterator) IterationStyle() wfapi.IterationStyle {
	return it.style
}

func (it *iterator) Producer() wfapi.Activity {
	return it.producer
}

func (it *iterator) Over() []eval.Parameter {
	return it.over
}

func (it *iterator) Variables() []eval.Parameter {
	return it.variables
}
