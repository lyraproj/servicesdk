package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/servicesdk/serviceapi"
)

type Reference interface {
	Step

	// Reference is the name of the activity that is referenced
	Reference() string
}

type reference struct {
	step
	referencedStep string
}

func MakeReference(name string, origin issue.Location, when Condition, input, output []serviceapi.Parameter, referencedStep string) Reference {
	return &reference{step{name, origin, when, input, output}, referencedStep}
}

func (s *reference) Label() string {
	return `reference ` + s.name
}

func (s *reference) Reference() string {
	return s.referencedStep
}
