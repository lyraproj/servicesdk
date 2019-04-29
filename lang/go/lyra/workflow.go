package lyra

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wf"
)

// Workflow groups several activities into one activity. Dependencies between the activities are determined
// by their input and output declarations
type Workflow struct {
	// Name of workflow. This field is mandatory
	Name string

	// When is a Condition in string form. Can be left empty
	When string

	// Input is an optional zero value of a struct or a pointer to a struct. The exported fields
	// of that struct defines the input of the workflow activity
	Input interface{}

	// Output is an optional zero value of a struct or a pointer to a struct. The exported fields
	// of that struct defines the output of the workflow activity
	Output interface{}

	// Activities is the slice of activities that are executed by this workflow
	Activities []Activity
}

func (w *Workflow) Resolve(c px.Context, pn string) wf.Activity {
	n := w.Name
	if n == `` {
		panic(px.Error(MissingRequiredField, issue.H{`type`: `Workflow`, `name`: `Name`}))
	}
	if pn != `` {
		n = pn + `::` + n
	}

	as := make([]wf.Activity, len(w.Activities))
	for i, a := range w.Activities {
		as[i] = a.Resolve(c, n)
	}
	return wf.MakeWorkflow(n, wf.Parse(w.When), ParametersFromGoStruct(c, w.Input), ParametersFromGoStruct(c, w.Output), as)
}
