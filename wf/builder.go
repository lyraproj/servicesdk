package wf

import (
	"strings"

	"github.com/lyraproj/servicesdk/serviceapi"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

var noParams = make([]serviceapi.Parameter, 0)

func LeafName(name string) string {
	names := strings.Split(name, `::`)
	return names[len(names)-1]
}

type Builder interface {
	Context() px.Context
	Build() Step
	Name(string)
	When(string)
	Parameters(...serviceapi.Parameter)
	Returns(...serviceapi.Parameter)
	QualifyName(childName string) string
	GetParameters() []serviceapi.Parameter
	GetName() string
	Parameter(name, typeName string) serviceapi.Parameter
}

type ChildBuilder interface {
	Builder
	StateHandler(func(StateHandlerBuilder))
	Resource(func(ResourceBuilder))
	Workflow(func(WorkflowBuilder))
	Action(func(ActionBuilder))
	Call(func(CallBuilder))
	AddChild(Builder)
	Iterator(func(IteratorBuilder))
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
	Over(px.Value)
	Variables(...serviceapi.Parameter)
	Into(into string)
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

type CallBuilder interface {
	Builder
	CallTo(string)
}

type WorkflowBuilder interface {
	ChildBuilder
}

func NewStateHandler(ctx px.Context, bf func(StateHandlerBuilder)) StateHandler {
	bld := &stateHandlerBuilder{builder: builder{ctx: ctx, when: Always, parameters: noParams, returns: noParams, origin: ctx.StackTop()}}
	bf(bld)
	return bld.Build().(StateHandler)
}

func NewIterator(ctx px.Context, bf func(IteratorBuilder)) Iterator {
	bld := &iteratorBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: Always, parameters: noParams, returns: noParams, origin: ctx.StackTop()}}}
	bf(bld)
	return bld.Build().(Iterator)
}

func NewResource(ctx px.Context, bf func(ResourceBuilder)) Resource {
	bld := &resourceBuilder{builder: builder{ctx: ctx, when: Always, parameters: noParams, returns: noParams, origin: ctx.StackTop()}}
	bf(bld)
	return bld.Build().(Resource)
}

func NewAction(ctx px.Context, bf func(ActionBuilder)) Action {
	bld := &actionBuilder{builder: builder{ctx: ctx, when: Always, parameters: noParams, returns: noParams, origin: ctx.StackTop()}}
	bf(bld)
	return bld.Build().(Action)
}

func NewCall(ctx px.Context, bf func(CallBuilder)) Call {
	bld := &callBuilder{builder: builder{ctx: ctx, when: Always, parameters: noParams, returns: noParams, origin: ctx.StackTop()}}
	bf(bld)
	return bld.Build().(Call)
}

func NewWorkflow(ctx px.Context, bf func(WorkflowBuilder)) Workflow {
	bld := &workflowBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: Always, parameters: noParams, returns: noParams, origin: ctx.StackTop()}}}
	bf(bld)
	return bld.Build().(Workflow)
}

type builder struct {
	ctx        px.Context
	origin     issue.Location
	name       string
	when       Condition
	parameters []serviceapi.Parameter
	returns    []serviceapi.Parameter
	parent     Builder
}

func (b *builder) Context() px.Context {
	return b.ctx
}

func (b *builder) Name(n string) {
	b.name = n
}

func (b *builder) When(w string) {
	if w == `` {
		b.when = Always
	} else {
		b.when = Parse(w)
	}
}

func (b *builder) validate() {
	if b.name == `` {
		panic(px.Error(StepNoName, issue.NoArgs))
	}
}

func (b *builder) Parameter(name, typeName string) serviceapi.Parameter {
	return serviceapi.NewParameter(name, ``, b.ctx.ParseType(typeName), nil)
}

func (b *builder) GetParameters() []serviceapi.Parameter {
	return b.parameters
}

func (b *builder) GetOrigin() issue.Location {
	return b.origin
}

func (b *builder) QualifyName(childName string) string {
	return b.GetName() + `::` + childName
}

func (b *builder) GetName() string {
	if b.parent != nil {
		return b.parent.QualifyName(b.name)
	}
	return b.name
}

func (b *builder) Parameters(parameters ...serviceapi.Parameter) {
	if len(b.parameters) == 0 {
		b.parameters = parameters
	} else {
		b.parameters = append(b.parameters, parameters...)
	}
}

func (b *builder) Returns(returns ...serviceapi.Parameter) {
	if len(b.returns) == 0 {
		b.returns = returns
	} else {
		b.returns = append(b.returns, returns...)
	}
}

type stateHandlerBuilder struct {
	builder
	api interface{}
}

func (b *stateHandlerBuilder) API(c interface{}) {
	b.api = c
}

func (b *stateHandlerBuilder) Build() Step {
	b.validate()
	return MakeStateHandler(b.GetName(), b.origin, b.when, b.parameters, b.returns, b.api)
}

type childBuilder struct {
	builder
	children []Step
}

func stateHandlerChild(b ChildBuilder, bld func(b StateHandlerBuilder)) {
	ab := &stateHandlerBuilder{builder: builder{parent: b, ctx: b.Context(), when: Always, parameters: noParams, returns: noParams, origin: b.Context().StackTop()}}
	bld(ab)
	b.AddChild(ab)
}

func resourceChild(b ChildBuilder, bld func(b ResourceBuilder)) {
	ab := &resourceBuilder{builder: builder{parent: b, ctx: b.Context(), when: Always, parameters: noParams, returns: noParams, origin: b.Context().StackTop()}}
	bld(ab)
	b.AddChild(ab)
}

func workflowChild(b ChildBuilder, bld func(b WorkflowBuilder)) {
	ab := &workflowBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.Context(), when: Always, parameters: noParams, returns: noParams, origin: b.Context().StackTop()}}}
	bld(ab)
	b.AddChild(ab)
}

func actionChild(b ChildBuilder, bld func(b ActionBuilder)) {
	ab := &actionBuilder{builder: builder{parent: b, ctx: b.Context(), when: Always, parameters: noParams, returns: noParams, origin: b.Context().StackTop()}}
	bld(ab)
	b.AddChild(ab)
}

func callChild(b ChildBuilder, bld func(b CallBuilder)) {
	ab := &callBuilder{builder: builder{parent: b, ctx: b.Context(), when: Always, parameters: noParams, returns: noParams, origin: b.Context().StackTop()}}
	bld(ab)
	b.AddChild(ab)
}

func (b *childBuilder) AddChild(child Builder) {
	b.children = append(b.children, child.Build())
}

type iteratorBuilder struct {
	childBuilder
	style     IterationStyle
	over      px.Value
	variables []serviceapi.Parameter
	into      string
}

func (b *iteratorBuilder) StateHandler(bld func(b StateHandlerBuilder)) {
	stateHandlerChild(b, bld)
}

func (b *iteratorBuilder) Resource(bld func(b ResourceBuilder)) {
	resourceChild(b, bld)
}

func (b *iteratorBuilder) Workflow(bld func(b WorkflowBuilder)) {
	workflowChild(b, bld)
}

func (b *iteratorBuilder) Action(bld func(b ActionBuilder)) {
	actionChild(b, bld)
}

func (b *iteratorBuilder) Call(bld func(b CallBuilder)) {
	callChild(b, bld)
}

func (b *iteratorBuilder) Iterator(bld func(b IteratorBuilder)) {
	ab := &iteratorBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.ctx, when: Always, parameters: noParams, returns: noParams}}}
	bld(ab)
	b.AddChild(ab)
}

func (b *iteratorBuilder) GetName() string {
	if b.name == `` {
		if len(b.children) != 1 {
			panic(`ouch`)
		}
		return b.children[0].Name()
	}
	return b.parent.QualifyName(b.name)
}

func (b *iteratorBuilder) QualifyName(childName string) string {
	if b.parent == nil {
		return childName
	}
	return b.parent.QualifyName(childName)
}

func (b *iteratorBuilder) Style(style IterationStyle) {
	b.style = style
}

func (b *iteratorBuilder) Over(over px.Value) {
	b.over = over
}

func (b *iteratorBuilder) Into(into string) {
	b.into = into
}

func (b *iteratorBuilder) Variables(variables ...serviceapi.Parameter) {
	if len(b.variables) == 0 {
		b.variables = variables
	} else {
		b.variables = append(b.variables, variables...)
	}
}

func (b *iteratorBuilder) Build() Step {
	b.validate()
	return MakeIterator(b.GetName(), b.origin, b.when, b.parameters, b.returns, b.style, b.children[0], b.over, b.variables, b.into)
}

func (b *iteratorBuilder) validate() {
	if len(b.children) != 1 {
		panic(px.Error(IteratorNotOneStep, issue.NoArgs))
	}
}

type resourceBuilder struct {
	builder
	state State
	extId string
}

func (b *resourceBuilder) Build() Step {
	b.validate()
	return MakeResource(b.GetName(), b.origin, b.when, b.parameters, b.returns, b.extId, b.state)
}

func (b *resourceBuilder) State(state State) {
	b.state = state
}

func (b *resourceBuilder) ExternalId(extId string) {
	b.extId = extId
}

// RegisterState registers a struct as a state. The state type is inferred from the
// struct
func (b *resourceBuilder) StateStruct(state interface{}) {
	/* TODO: Fix this b.state = newGoState(pt.(px.ObjectType), rv)
	rv := reflect.ValueOf(state)
	rt := rv.Type()
	pt, ok := b.ctx.ImplementationRegistry().ReflectedToType(rt)
	if !ok {
		pt = b.ctx.Reflector().TypeFromReflect(b.GetName(), nil, rt)
	}
	*/
}

type workflowBuilder struct {
	childBuilder
}

func (b *workflowBuilder) Build() Step {
	b.validate()
	return MakeWorkflow(b.GetName(), b.origin, b.when, b.parameters, b.returns, b.children)
}

func (b *workflowBuilder) StateHandler(bld func(b StateHandlerBuilder)) {
	stateHandlerChild(b, bld)
}

func (b *workflowBuilder) Resource(bld func(b ResourceBuilder)) {
	resourceChild(b, bld)
}

func (b *workflowBuilder) Workflow(bld func(b WorkflowBuilder)) {
	workflowChild(b, bld)
}

func (b *workflowBuilder) Action(bld func(b ActionBuilder)) {
	actionChild(b, bld)
}

func (b *workflowBuilder) Call(bld func(b CallBuilder)) {
	callChild(b, bld)
}

func (b *workflowBuilder) Iterator(bld func(b IteratorBuilder)) {
	ab := &iteratorBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.ctx, when: Always, parameters: noParams, returns: noParams}}}
	bld(ab)
	b.AddChild(ab)
}

type actionBuilder struct {
	builder
	function interface{}
}

func (b *actionBuilder) Build() Step {
	b.validate()
	return MakeAction(b.GetName(), b.origin, b.when, b.parameters, b.returns, b.function)
}

func (b *actionBuilder) Doer(d interface{}) {
	b.function = d
}

type callBuilder struct {
	builder
	calledStep string
}

func (b *callBuilder) Build() Step {
	b.validate()
	return MakeCall(b.GetName(), b.origin, b.when, b.parameters, b.returns, b.calledStep)
}

func (b *callBuilder) CallTo(calledStep string) {
	b.calledStep = calledStep
}
