package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/servicesdk/serviceapi"
)

type Call interface {
	Step

	// Call is the name of the activity that is called
	Call() string
}

type call struct {
	step
	calledStep string
}

func MakeCall(name string, origin issue.Location, when Condition, input, output []serviceapi.Parameter, calledStep string) Call {
	return &call{step{name, origin, when, input, output}, calledStep}
}

func (s *call) Label() string {
	return `call ` + s.name
}

func (s *call) Call() string {
	return s.calledStep
}
