package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/condition"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/wfapi"
	"reflect"
)

var noParams = []px.Parameter{}

func init() {
	wfapi.NewStateHandler = func(ctx px.Context, bf func(wfapi.StateHandlerBuilder)) wfapi.StateHandler {
		bld := &stateHandlerBuilder{builder: builder{ctx: ctx, when: condition.Always, input: noParams, output: noParams}}
		bf(bld)
		return bld.Build().(wfapi.StateHandler)
	}

	wfapi.NewIterator = func(ctx px.Context, bf func(wfapi.IteratorBuilder)) wfapi.Iterator {
		bld := &iteratorBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: condition.Always, input: noParams, output: noParams}}}
		bf(bld)
		return bld.Build().(wfapi.Iterator)
	}

	wfapi.NewResource = func(ctx px.Context, bf func(wfapi.ResourceBuilder)) wfapi.Resource {
		bld := &resourceBuilder{builder: builder{ctx: ctx, when: condition.Always, input: noParams, output: noParams}}
		bf(bld)
		return bld.Build().(wfapi.Resource)
	}

	wfapi.NewAction = func(ctx px.Context, bf func(wfapi.ActionBuilder)) wfapi.Action {
		bld := &actionBuilder{builder: builder{ctx: ctx, when: condition.Always, input: noParams, output: noParams}}
		bf(bld)
		return bld.Build().(wfapi.Action)
	}

	wfapi.NewWorkflow = func(ctx px.Context, bf func(wfapi.WorkflowBuilder)) wfapi.Workflow {
		bld := &workflowBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: condition.Always, input: noParams, output: noParams}}}
		bf(bld)
		return bld.Build().(wfapi.Workflow)
	}
}

type builder struct {
	ctx    px.Context
	name   string
	when   wfapi.Condition
	input  []px.Parameter
	output []px.Parameter
	parent wfapi.Builder
}

func (b *builder) Context() px.Context {
	return b.ctx
}

func (b *builder) Name(n string) {
	b.name = n
}

func (b *builder) When(w string) {
	if w == `` {
		b.when = condition.Always
	} else {
		b.when = condition.Parse(w)
	}
}

func (b *builder) validate() {
	if b.name == `` {
		panic(px.Error(wfapi.WF_ACTIVITY_NO_NAME, issue.NO_ARGS))
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

func (b *stateHandlerBuilder) Build() wfapi.Activity {
	b.validate()
	return NewStateHandler(b.GetName(), b.when, b.input, b.output, b.api)
}

type childBuilder struct {
	builder
	children []wfapi.Activity
}

func stateHandlerChild(b wfapi.ChildBuilder, bld func(b wfapi.StateHandlerBuilder)) {
	ab := &stateHandlerBuilder{builder: builder{parent: b, ctx: b.Context(), when: condition.Always, input: noParams, output: noParams}}
	bld(ab)
	b.AddChild(ab)
}

func resourceChild(b wfapi.ChildBuilder, bld func(b wfapi.ResourceBuilder)) {
	ab := &resourceBuilder{builder: builder{parent: b, ctx: b.Context(), when: condition.Always, input: noParams, output: noParams}}
	bld(ab)
	b.AddChild(ab)
}

func workflowChild(b wfapi.ChildBuilder, bld func(b wfapi.WorkflowBuilder)) {
	ab := &workflowBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.Context(), when: condition.Always, input: noParams, output: noParams}}}
	bld(ab)
	b.AddChild(ab)
}

func actionChild(b wfapi.ChildBuilder, bld func(b wfapi.ActionBuilder)) {
	ab := &actionBuilder{builder: builder{parent: b, ctx: b.Context(), when: condition.Always, input: noParams, output: noParams}}
	bld(ab)
	b.AddChild(ab)
}

func (b *childBuilder) AddChild(child wfapi.Builder) {
	b.children = append(b.children, child.Build())
}

type iteratorBuilder struct {
	childBuilder
	style     wfapi.IterationStyle
	over      []px.Parameter
	variables []px.Parameter
}

func (b *iteratorBuilder) StateHandler(bld func(b wfapi.StateHandlerBuilder)) {
	stateHandlerChild(b, bld)
}

func (b *iteratorBuilder) Resource(bld func(b wfapi.ResourceBuilder)) {
	resourceChild(b, bld)
}

func (b *iteratorBuilder) Workflow(bld func(b wfapi.WorkflowBuilder)) {
	workflowChild(b, bld)
}

func (b *iteratorBuilder) Action(bld func(b wfapi.ActionBuilder)) {
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

func (b *iteratorBuilder) Style(style wfapi.IterationStyle) {
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

func (b *iteratorBuilder) Build() wfapi.Activity {
	b.validate()
	return NewIterator(b.GetName(), b.when, b.input, b.output, b.style, b.children[0], b.over, b.variables)
}

func (b *iteratorBuilder) validate() {
	if len(b.children) != 1 {
		panic(px.Error(wfapi.WF_ITERATOR_NOT_ONE_ACTIVITY, issue.NO_ARGS))
	}
}

type resourceBuilder struct {
	builder
	state wfapi.State
	extId string
}

func (b *resourceBuilder) Build() wfapi.Activity {
	b.validate()
	return NewResource(b.GetName(), b.when, b.input, b.output, b.extId, b.state)
}

func (b *resourceBuilder) State(state wfapi.State) {
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
	b.state = service.NewGoState(pt.(px.ObjectType), rv)
}

type workflowBuilder struct {
	childBuilder
}

func (b *workflowBuilder) Build() wfapi.Activity {
	b.validate()
	return NewWorkflow(b.GetName(), b.when, b.input, b.output, b.children)
}

func (b *workflowBuilder) StateHandler(bld func(b wfapi.StateHandlerBuilder)) {
	stateHandlerChild(b, bld)
}

func (b *workflowBuilder) Resource(bld func(b wfapi.ResourceBuilder)) {
	resourceChild(b, bld)
}

func (b *workflowBuilder) Workflow(bld func(b wfapi.WorkflowBuilder)) {
	workflowChild(b, bld)
}

func (b *workflowBuilder) Action(bld func(b wfapi.ActionBuilder)) {
	actionChild(b, bld)
}

func (b *workflowBuilder) Iterator(bld func(b wfapi.IteratorBuilder)) {
	ab := &iteratorBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.ctx, when: condition.Always, input: noParams, output: noParams}}}
	bld(ab)
	b.AddChild(ab)
}

type actionBuilder struct {
	builder
	function interface{}
}

func (b *actionBuilder) Build() wfapi.Activity {
	b.validate()
	return NewAction(b.GetName(), b.when, b.input, b.output, b.function)
}

func (b *actionBuilder) Doer(d interface{}) {
	b.function = d
}
