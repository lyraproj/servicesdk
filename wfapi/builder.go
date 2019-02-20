package wfapi

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"strings"
)

func LeafName(name string) string {
	names := strings.Split(name, `::`)
	return names[len(names)-1]
}

type Builder interface {
	Context() eval.Context
	Build() Activity
	Name(string)
	When(string)
	Input(...eval.Parameter)
	Output(...eval.Parameter)
	QualifyName(childName string) string
	GetInput() []eval.Parameter
	GetName() string
	Parameter(name, typeName string) eval.Parameter
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
	Over(...eval.Parameter)
	Variables(...eval.Parameter)
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

var NewStateHandler func(eval.Context, func(StateHandlerBuilder)) StateHandler
var NewIterator func(eval.Context, func(IteratorBuilder)) Iterator
var NewResource func(eval.Context, func(ResourceBuilder)) Resource
var NewAction func(eval.Context, func(ActionBuilder)) Action
var NewWorkflow func(eval.Context, func(WorkflowBuilder)) Workflow
