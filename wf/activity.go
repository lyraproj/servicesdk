package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

type Resolvable interface {
	// Resolve recreates the instance using the given Context and returns something new
	Resolve(c px.Context) interface{}
}

// An Activity of a Workflow. The workflow is an Activity in itself and can be used in
// another Workflow.
type Activity interface {
	issue.Labeled

	// When returns an optional Condition that controls whether or not this activity participates
	// in the workflow.
	When() Condition

	// Name returns the fully qualified name of the Activity
	Name() string

	// Input returns the input requirements for the Activity
	Input() []px.Parameter

	// Output returns the definition of that this Activity will produce
	Output() []px.Parameter

	// Resolve resolves the activity internals using the given Context
	Resolve(c px.Context)
}

type activity struct {
	name   string
	when   Condition
	input  []px.Parameter
	output []px.Parameter
}

func (a *activity) When() Condition {
	return a.when
}

func (a *activity) Name() string {
	return a.name
}

func (a *activity) Input() []px.Parameter {
	return a.input
}

func (a *activity) Output() []px.Parameter {
	return a.output
}

func (a *activity) Resolve(px.Context) {
}
