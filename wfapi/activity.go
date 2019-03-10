package wfapi

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

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
}
