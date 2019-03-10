package wfapi

import (
	"github.com/lyraproj/pcore/px"
	"strings"
)

func LeafName(name string) string {
	names := strings.Split(name, `::`)
	return names[len(names)-1]
}

type Builder interface {
	Context() px.Context
	Build() Activity
	Name(string)
	When(string)
	Input(...px.Parameter)
	Output(...px.Parameter)
	QualifyName(childName string) string
	GetInput() []px.Parameter
	GetName() string
	Parameter(name, typeName string) px.Parameter
}

type ChildBuilder interface {
	Builder
	StateHandler(func(StateHandlerBuilder))
	Resource(func(ResourceBuilder))
	Workflow(func(WorkflowBuilder))
	Action(func(ActionBuilder))
	AddChild(Builder)
}

type APIBuilder interface {
	Builder
	API(interface{})
}

type StateHandlerBuilder interface {
	Builder
	API(interface{})
}

type IteratorBuilder interface {
	ChildBuilder
	Style(IterationStyle)
	Over(...px.Parameter)
	Variables(...px.Parameter)
}

type ResourceBuilder interface {
	Builder
	ExternalId(extId string)
	State(state State)
	StateStruct(state interface{})
}

type ActionBuilder interface {
	Builder
	Doer(interface{})
}

type WorkflowBuilder interface {
	ChildBuilder
	Iterator(func(IteratorBuilder))
}

var NewStateHandler func(px.Context, func(StateHandlerBuilder)) StateHandler
var NewIterator func(px.Context, func(IteratorBuilder)) Iterator
var NewResource func(px.Context, func(ResourceBuilder)) Resource
var NewAction func(px.Context, func(ActionBuilder)) Action
var NewWorkflow func(px.Context, func(WorkflowBuilder)) Workflow
