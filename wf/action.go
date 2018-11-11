package wf

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-servicesdk/wfapi"
)

type action struct {
	activity
	crd wfapi.CRD
}

func NewAction(name string, when wfapi.Condition, input, output []eval.Parameter, crd wfapi.CRD) wfapi.Action {
	return &action{activity{name, when, input, output}, crd}
}

func (a *action) Label() string {
	return `action ` + a.name
}

func (a *action) Interface() wfapi.CRD {
	return a.crd
}
