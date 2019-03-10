package wf

import (
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wfapi"
)

type workflow struct {
	activity
	activities []wfapi.Activity
}

func NewWorkflow(name string, when wfapi.Condition, input, output []px.Parameter, activities []wfapi.Activity) wfapi.Workflow {
	return &workflow{activity{name, when, input, output}, activities}
}

func (w *workflow) Label() string {
	return `workflow ` + w.name
}

func (w *workflow) Activities() []wfapi.Activity {
	return w.activities
}
