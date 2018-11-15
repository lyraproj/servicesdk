package wf

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-servicesdk/condition"
	"github.com/puppetlabs/go-servicesdk/service"
	"github.com/puppetlabs/go-servicesdk/wfapi"
	"reflect"
)

func init() {
	wfapi.NewAction = func(ctx eval.Context, bf func(wfapi.ActionBuilder)) wfapi.Action {
		bld := &actionBuilder{builder: builder{ctx: ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
		bf(bld)
		return bld.build()
	}

	wfapi.NewIterator = func(ctx eval.Context, bf func(wfapi.IteratorBuilder)) wfapi.Iterator {
		bld := &iteratorBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}}
		bf(bld)
		return bld.build()
	}

	wfapi.NewResource = func(ctx eval.Context, bf func(wfapi.ResourceBuilder)) wfapi.Resource {
		bld := &resourceBuilder{builder: builder{ctx: ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
		bf(bld)
		return bld.build()
	}

	wfapi.NewStateless = func(ctx eval.Context, bf func(wfapi.StatelessBuilder)) wfapi.Stateless {
		bld := &statelessBuilder{builder: builder{ctx: ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
		bf(bld)
		return bld.build()
	}

	wfapi.NewWorkflow = func(ctx eval.Context, bf func(wfapi.WorkflowBuilder)) wfapi.Workflow {
		bld := &workflowBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}}
		bf(bld)
		return bld.build()
	}
}

type builder struct {
	ctx eval.Context
	name string
	when wfapi.Condition
	input []eval.Parameter
	output []eval.Parameter
	parent wfapi.Builder
}

func (b *builder) Context() eval.Context {
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
		panic(eval.Error(wfapi.WF_ACTIVITY_NO_NAME, issue.NO_ARGS))
	}
}

func (b *builder) Parameter(name, typeName string) eval.Parameter {
	return impl.NewParameter(name, b.ctx.ParseType2(typeName), nil, false)
}

func (b *builder) GetInput() []eval.Parameter {
	return b.input
}

func (b *builder) GetName() string {
	if b.parent != nil {
		return b.parent.GetName() + `::` + b.name
	}
	return b.name
}

func (b *builder) Input(input ...eval.Parameter) {
	if len(b.input) == 0 {
		b.input = input
	} else {
		b.input = append(b.input, input...)
	}
}

func (b *builder) Output(output ...eval.Parameter) {
	if len(b.output) == 0 {
		b.output = output
	} else {
		b.output = append(b.output, output...)
	}
}

type actionBuilder struct {
	builder
	api interface{}
}

func (b *actionBuilder) API(c interface{}) {
	b.api = c
}

func (b *actionBuilder) build() wfapi.Action {
	b.validate()
	return NewAction(b.GetName(), b.when, b.input, b.output, b.api)
}

type childBuilder struct {
	builder
	children []wfapi.Activity
}

func (b *childBuilder) Action(bld func(b wfapi.ActionBuilder)) {
	ab := &actionBuilder{builder: builder{parent: b, ctx: b.ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
	bld(ab)
	b.children = append(b.children, ab.build())
}

func (b *childBuilder) Resource(bld func(b wfapi.ResourceBuilder)) {
	ab := &resourceBuilder{builder: builder{parent: b, ctx: b.ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
	bld(ab)
	b.children = append(b.children, ab.build())
}

func (b *childBuilder) Workflow(bld func(b wfapi.WorkflowBuilder)) {
	ab := &workflowBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}}
	bld(ab)
	b.children = append(b.children, ab.build())
}

func (b *childBuilder) Stateless(bld func(b wfapi.StatelessBuilder)) {
	ab := &statelessBuilder{builder: builder{parent: b, ctx: b.ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
	bld(ab)
	b.children = append(b.children, ab.build())
}


type iteratorBuilder struct {
	childBuilder
	style wfapi.IterationStyle
	over []eval.Parameter
	variables []eval.Parameter
}

func (b *iteratorBuilder) Style(style wfapi.IterationStyle) {
	b.style = style
}

func (b *iteratorBuilder) Over(over ...eval.Parameter) {
	if len(b.over) == 0 {
		b.over = over
	} else {
		b.over = append(b.over, over...)
	}
}

func (b *iteratorBuilder) Variables(variables ...eval.Parameter) {
	if len(b.variables) == 0 {
		b.variables = variables
	} else {
		b.variables = append(b.variables, variables...)
	}
}

func (b *iteratorBuilder) build() wfapi.Iterator {
	b.validate()
	return NewIterator(b.GetName(), b.when, b.input, b.output, b.style, b.children[0], b.over, b.variables)
}

func (b *iteratorBuilder) validate() {
	if len(b.children) != 1 {
		panic(eval.Error(wfapi.WF_ITERATOR_NOT_ONE_ACTIVITY, issue.NO_ARGS))
	}
	if b.name == `` {
		b.name = b.children[0].Name()
	}
	b.builder.validate()
}

type resourceBuilder struct {
	builder
	state wfapi.State
}

func (b *resourceBuilder) build() wfapi.Resource {
	b.validate()
	return NewResource(b.GetName(), b.when, b.input, b.output, b.state)
}

func (b *resourceBuilder) State(state wfapi.State) {
	b.state = state
}

// RegisterState registers a struct as a state. The state type is inferred from the
// struct
func (b *resourceBuilder) StateStruct(state interface{}) {
	rv := reflect.ValueOf(state)
	rt := rv.Type()
	pt, ok := b.ctx.ImplementationRegistry().ReflectedToType(rt)
	if !ok {
		pt = b.ctx.Reflector().ObjectTypeFromReflect(b.GetName(), nil, rt)
	}
	b.state = service.NewGoState(pt.(eval.ObjectType), rv)
}

type workflowBuilder struct {
	childBuilder
}

func (b *workflowBuilder) build() wfapi.Workflow {
	b.validate()
	return NewWorkflow(b.GetName(), b.when, b.input, b.output, b.children)
}

func (b *workflowBuilder) Iterator(bld func(b wfapi.IteratorBuilder)) {
	ab := &iteratorBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}}
	bld(ab)
	b.children = append(b.children, ab.build())
}

type statelessBuilder struct {
	builder
	doer wfapi.Doer
}

func (b *statelessBuilder) build() wfapi.Stateless {
	b.validate()
	return NewStateless(b.GetName(), b.when, b.input, b.output, b.doer)
}

func (b *statelessBuilder) Doer(d wfapi.Doer) {
	b.doer = d
}
