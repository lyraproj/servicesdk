package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

// An Step of a Workflow. The workflow is an Step in itself and can be used in
// another Workflow.
type Step interface {
	issue.Labeled

	Origin() issue.Location

	// When returns an optional Condition that controls whether or not this step participates
	// in the workflow.
	When() Condition

	// Name returns the fully qualified name of the Step
	Name() string

	// Parameters returns the parameters requirements for the Step
	Parameters() []px.Parameter

	// Returns returns the definition of that this Step will produce
	Returns() []px.Parameter
}

type step struct {
	name       string
	origin     issue.Location
	when       Condition
	parameters []px.Parameter
	returns    []px.Parameter
}

func (a *step) When() Condition {
	return a.when
}

func (a *step) Name() string {
	return a.name
}

func (a *step) Origin() issue.Location {
	return a.origin
}

func (a *step) Parameters() []px.Parameter {
	return a.parameters
}

func (a *step) Returns() []px.Parameter {
	return a.returns
}

func (a *step) Resolve(px.Context) {
}
