package wf

import (
	"github.com/lyraproj/pcore/px"
)

type Workflow interface {
	Step

	Steps() []Step
}

type workflow struct {
	step
	steps []Step
}

func MakeWorkflow(name string, when Condition, parameters, returns []px.Parameter, steps []Step) Workflow {
	return &workflow{step{name, when, parameters, returns}, steps}
}

func (w *workflow) Label() string {
	return `workflow ` + w.name
}

func (w *workflow) Steps() []Step {
	return w.steps
}
