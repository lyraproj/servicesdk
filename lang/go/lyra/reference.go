package lyra

import (
	"reflect"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wf"
)

// Reference is a reference to an external loadable step.
type Reference struct {
	// When is a Condition in string form. Can be left empty
	When string

	// Parameters is an optional zero value of a struct or a pointer to a struct. The exported fields
	// of that struct defines the parameters of the reference step.
	//
	// Parameters on a reference should mainly be used to rename the parameters of the referenced step
	// using parameter aliases.
	//
	// If the struct is not provided, then the parameters of the referenced step will be used verbatim.
	Parameters interface{}

	// Return on a reference should mainly be used to rename the values returned from the referenced step.
	Return interface{}

	// StepName is the name of the referenced step
	StepName string
}

func (r *Reference) Resolve(c px.Context, n string) wf.Step {
	var parameters, returns []px.Parameter
	if r.Parameters != nil {
		parameters = paramsFromStruct(c, reflect.TypeOf(r.Parameters), issue.FirstToLower)
	}
	if r.Return != nil {
		returns = paramsFromStruct(c, reflect.TypeOf(r.Return), issue.FirstToLower)
	}
	return wf.MakeReference(
		n, wf.Parse(r.When), parameters, returns, r.StepName)
}
