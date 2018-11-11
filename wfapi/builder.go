package wfapi

import "github.com/puppetlabs/go-evaluator/eval"

type Builder interface {
	Name(n string)
	When(w string)
	Input(name, typeName string)
	Output(name, typeName string)
}

type ActionBuilder interface {
	Builder
	CRD(c CRD)
}

type ResourceBuilder interface {
	Builder
	State(o interface{})
}

type WorkflowBuilder interface {
	Builder
	Action(bld func(b ActionBuilder))
	Resource(bld func(b ResourceBuilder))
	Workflow(bld func(b WorkflowBuilder))
	Stateless(bld func(b StatelessBuilder))
}

type StatelessBuilder interface {
	Builder
	Doer(d Doer)
}

var NewWorkflow func(eval.Context, func(WorkflowBuilder)) Workflow
