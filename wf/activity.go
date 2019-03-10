package wf

import (
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wfapi"
)

type activity struct {
	name   string
	when   wfapi.Condition
	input  []px.Parameter
	output []px.Parameter
}

func (a *activity) When() wfapi.Condition {
	return a.when
}

func (a *activity) Name() string {
	return a.name
}

func (a *activity) Input() []px.Parameter {
	return a.input
}

func (a *activity) Output() []px.Parameter {
	return a.output
}
