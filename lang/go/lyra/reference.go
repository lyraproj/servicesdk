package lyra

import (
	"reflect"

	"github.com/lyraproj/servicesdk/serviceapi"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wf"
)

// Call is a call to an external loadable step.
type Call struct {
	// When is a Condition in string form. Can be left empty
	When string

	// Parameters is an optional zero value of a struct or a pointer to a struct. The exported fields
	// of that struct defines the parameters of the call step.
	//
	// Parameters on a call should mainly be used to rename the parameters of the called step
	// using parameter aliases.
	//
	// If the struct is not provided, then the parameters of the called step will be used verbatim.
	Parameters interface{}

	// Return on a call should mainly be used to rename the values returned from the called step.
	Return interface{}

	// StepName is the name of the called step
	StepName string
}

func (r *Call) Resolve(c px.Context, n string, loc issue.Location) wf.Step {
	var parameters, returns []serviceapi.Parameter
	if r.Parameters != nil {
		parameters = paramsFromStruct(c, reflect.TypeOf(r.Parameters), issue.FirstToLower)
	}
	if r.Return != nil {
		returns = paramsFromStruct(c, reflect.TypeOf(r.Return), issue.FirstToLower)
	}
	return wf.MakeCall(
		n, loc, wf.Parse(r.When), parameters, returns, r.StepName)
}
