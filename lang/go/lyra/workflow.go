package lyra

import (
	"sort"

	"github.com/lyraproj/issue/issue"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wf"
)

// Workflow groups several steps into one step. Dependencies between the steps are determined
// by their parameters and returns declarations
type Workflow struct {
	// When is a Condition in string form. Can be left empty
	When string

	// Parameters is an optional zero value of a struct or a pointer to a struct. The exported fields
	// of that struct defines the parameters of the workflow step
	Parameters interface{}

	// Return is an optional zero value of a struct or a pointer to a struct. The exported fields
	// of that struct defines the returns of the workflow step
	Return interface{}

	// Steps is the slice of steps that are executed by this workflow
	Steps map[string]Step
}

func (w *Workflow) Resolve(c px.Context, n string, loc issue.Location) wf.Step {
	as := make([]wf.Step, 0, len(w.Steps))
	for k, a := range w.Steps {
		as = append(as, a.Resolve(c, n+`::`+k, loc))
	}
	sort.Slice(as, func(i, j int) bool {
		return as[i].Name() < as[j].Name()
	})
	return wf.MakeWorkflow(
		n, loc, wf.Parse(w.When), ParametersFromGoStruct(c, w.Parameters), ParametersFromGoStruct(c, w.Return), as)
}
