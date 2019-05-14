package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/servicesdk/serviceapi"
)

type Workflow interface {
	Step

	Steps() []Step
}

type workflow struct {
	step
	steps []Step
}

func MakeWorkflow(name string, origin issue.Location, when Condition, parameters, returns []serviceapi.Parameter, steps []Step) Workflow {
	return &workflow{step{name, origin, when, parameters, returns}, steps}
}

func (w *workflow) Label() string {
	return `workflow ` + w.name
}

func (w *workflow) Steps() []Step {
	return w.steps
}
