package wf

import (
	"github.com/lyraproj/pcore/px"
)

type Workflow interface {
	Activity

	Activities() []Activity
}

type workflow struct {
	activity
	activities []Activity
}

func MakeWorkflow(name string, when Condition, input, output []px.Parameter, activities []Activity) Workflow {
	return &workflow{activity{name, when, input, output}, activities}
}

func (w *workflow) Label() string {
	return `workflow ` + w.name
}

func (w *workflow) Activities() []Activity {
	return w.activities
}

func (w *workflow) Resolve(c px.Context) {
	for _, a := range w.activities {
		a.Resolve(c)
	}
}
