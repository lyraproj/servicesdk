package wf

import (
	"reflect"
	"strings"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

var noParams = make([]px.Parameter, 0)

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

func NewStateHandler(ctx px.Context, bf func(StateHandlerBuilder)) StateHandler {
	bld := &stateHandlerBuilder{builder: builder{ctx: ctx, when: Always, input: noParams, output: noParams}}
	bf(bld)
	return bld.Build().(StateHandler)
}

func NewIterator(ctx px.Context, bf func(IteratorBuilder)) Iterator {
	bld := &iteratorBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: Always, input: noParams, output: noParams}}}
	bf(bld)
	return bld.Build().(Iterator)
}

func NewResource(ctx px.Context, bf func(ResourceBuilder)) Resource {
	bld := &resourceBuilder{builder: builder{ctx: ctx, when: Always, input: noParams, output: noParams}}
	bf(bld)
	return bld.Build().(Resource)
}

func NewAction(ctx px.Context, bf func(ActionBuilder)) Action {
	bld := &actionBuilder{builder: builder{ctx: ctx, when: Always, input: noParams, output: noParams}}
	bf(bld)
	return bld.Build().(Action)
}

func NewWorkflow(ctx px.Context, bf func(WorkflowBuilder)) Workflow {
	bld := &workflowBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: Always, input: noParams, output: noParams}}}
	bf(bld)
	return bld.Build().(Workflow)
}

type builder struct {
	ctx    px.Context
	name   string
	when   Condition
	input  []px.Parameter
	output []px.Parameter
	parent Builder
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
		panic(px.Error(ActivityNoName, issue.NoArgs))
	}
}

func (b *builder) Parameter(name, typeName string) px.Parameter {
	return px.NewParameter(name, b.ctx.ParseType(typeName), nil, false)
}

func (b *builder) GetInput() []px.Parameter {
	return b.input
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

func (b *builder) Input(input ...px.Parameter) {
	if len(b.input) == 0 {
		b.input = input
	} else {
		b.input = append(b.input, input...)
	}
}

func (b *builder) Output(output ...px.Parameter) {
	if len(b.output) == 0 {
		b.output = output
	} else {
		b.output = append(b.output, output...)
	}
}

type stateHandlerBuilder struct {
	builder
	api interface{}
}

func (b *stateHandlerBuilder) API(c interface{}) {
	b.api = c
}

func (b *stateHandlerBuilder) Build() Activity {
	b.validate()
	return MakeStateHandler(b.GetName(), b.when, b.input, b.output, b.api)
}

type childBuilder struct {
	builder
	children []Activity
}

func stateHandlerChild(b ChildBuilder, bld func(b StateHandlerBuilder)) {
	ab := &stateHandlerBuilder{builder: builder{parent: b, ctx: b.Context(), when: Always, input: noParams, output: noParams}}
	bld(ab)
	b.AddChild(ab)
}

func resourceChild(b ChildBuilder, bld func(b ResourceBuilder)) {
	ab := &resourceBuilder{builder: builder{parent: b, ctx: b.Context(), when: Always, input: noParams, output: noParams}}
	bld(ab)
	b.AddChild(ab)
}

func workflowChild(b ChildBuilder, bld func(b WorkflowBuilder)) {
	ab := &workflowBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.Context(), when: Always, input: noParams, output: noParams}}}
	bld(ab)
	b.AddChild(ab)
}

func actionChild(b ChildBuilder, bld func(b ActionBuilder)) {
	ab := &actionBuilder{builder: builder{parent: b, ctx: b.Context(), when: Always, input: noParams, output: noParams}}
	bld(ab)
	b.AddChild(ab)
}

func (b *childBuilder) AddChild(child Builder) {
	b.children = append(b.children, child.Build())
}

type iteratorBuilder struct {
	childBuilder
	style     IterationStyle
	over      []px.Parameter
	variables []px.Parameter
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

func (b *iteratorBuilder) Over(over ...px.Parameter) {
	if len(b.over) == 0 {
		b.over = over
	} else {
		b.over = append(b.over, over...)
	}
}

func (b *iteratorBuilder) Variables(variables ...px.Parameter) {
	if len(b.variables) == 0 {
		b.variables = variables
	} else {
		b.variables = append(b.variables, variables...)
	}
}

func (b *iteratorBuilder) Build() Activity {
	b.validate()
	return MakeIterator(b.GetName(), b.when, b.input, b.output, b.style, b.children[0], b.over, b.variables)
}

func (b *iteratorBuilder) validate() {
	if len(b.children) != 1 {
		panic(px.Error(IteratorNotOneActivity, issue.NoArgs))
	}
}

type resourceBuilder struct {
	builder
	state State
	extId string
}

func (b *resourceBuilder) Build() Activity {
	b.validate()
	return MakeResource(b.GetName(), b.when, b.input, b.output, b.extId, b.state)
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
	rv := reflect.ValueOf(state)
	rt := rv.Type()
	pt, ok := b.ctx.ImplementationRegistry().ReflectedToType(rt)
	if !ok {
		pt = b.ctx.Reflector().TypeFromReflect(b.GetName(), nil, rt)
	}
	b.state = NewGoState(pt.(px.ObjectType), rv)
}

type workflowBuilder struct {
	childBuilder
}

func (b *workflowBuilder) Build() Activity {
	b.validate()
	return MakeWorkflow(b.GetName(), b.when, b.input, b.output, b.children)
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

func (b *workflowBuilder) Iterator(bld func(b IteratorBuilder)) {
	ab := &iteratorBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.ctx, when: Always, input: noParams, output: noParams}}}
	bld(ab)
	b.AddChild(ab)
}

type actionBuilder struct {
	builder
	function interface{}
}

func (b *actionBuilder) Build() Activity {
	b.validate()
	return MakeAction(b.GetName(), b.when, b.input, b.output, b.function)
}

func (b *actionBuilder) Doer(d interface{}) {
	b.function = d
}
