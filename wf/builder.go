package wf

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/impl"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/servicesdk/condition"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/wfapi"
	"reflect"
)

func init() {
	wfapi.NewAction = func(ctx eval.Context, bf func(wfapi.ActionBuilder)) wfapi.Action {
		bld := &actionBuilder{builder: builder{ctx: ctx, when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}
		bf(bld)
		return bld.Build().(wfapi.Action)
	}

	wfapi.NewIterator = func(ctx eval.Context, bf func(wfapi.IteratorBuilder)) wfapi.Iterator {
		bld := &iteratorBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}}
		bf(bld)
		return bld.Build().(wfapi.Iterator)
	}

	wfapi.NewResource = func(ctx eval.Context, bf func(wfapi.ResourceBuilder)) wfapi.Resource {
		bld := &resourceBuilder{builder: builder{ctx: ctx, when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}
		bf(bld)
		return bld.Build().(wfapi.Resource)
	}

	wfapi.NewStateless = func(ctx eval.Context, bf func(wfapi.StatelessBuilder)) wfapi.Stateless {
		bld := &statelessBuilder{builder: builder{ctx: ctx, when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}
		bf(bld)
		return bld.Build().(wfapi.Stateless)
	}

	wfapi.NewWorkflow = func(ctx eval.Context, bf func(wfapi.WorkflowBuilder)) wfapi.Workflow {
		bld := &workflowBuilder{childBuilder: childBuilder{builder: builder{ctx: ctx, when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}}
		bf(bld)
		return bld.Build().(wfapi.Workflow)
	}
}

type builder struct {
	ctx    eval.Context
	name   string
	when   wfapi.Condition
	input  []eval.Parameter
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

func (b *builder) QualifyName(childName string) string {
		return b.GetName() + `::` + childName
}

func (b *builder) GetName() string {
	if b.parent != nil {
		return b.parent.QualifyName(b.name)
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

func (b *actionBuilder) Build() wfapi.Activity {
	b.validate()
	return NewAction(b.GetName(), b.when, b.input, b.output, b.api)
}

type childBuilder struct {
	builder
	children []wfapi.Activity
}

func actionChild(b wfapi.ChildBuilder, bld func(b wfapi.ActionBuilder)) {
	ab := &actionBuilder{builder: builder{parent: b, ctx: b.Context(), when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}
	bld(ab)
	b.AddChild(ab)
}

func resourceChild(b wfapi.ChildBuilder, bld func(b wfapi.ResourceBuilder)) {
	ab := &resourceBuilder{builder: builder{parent: b, ctx: b.Context(), when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}
	bld(ab)
	b.AddChild(ab)
}

func workflowChild(b wfapi.ChildBuilder, bld func(b wfapi.WorkflowBuilder)) {
	ab := &workflowBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.Context(), when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}}
	bld(ab)
	b.AddChild(ab)
}

func statelessChild(b wfapi.ChildBuilder, bld func(b wfapi.StatelessBuilder)) {
	ab := &statelessBuilder{builder: builder{parent: b, ctx: b.Context(), when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}
	bld(ab)
	b.AddChild(ab)
}

func (b *childBuilder) 	AddChild(child wfapi.Builder) {
	b.children = append(b.children, child.Build())
}

type iteratorBuilder struct {
	childBuilder
	style     wfapi.IterationStyle
	over      []eval.Parameter
	variables []eval.Parameter
}

func (b *iteratorBuilder) Action(bld func(b wfapi.ActionBuilder)) {
	actionChild(b, bld)
}

func (b *iteratorBuilder) Resource(bld func(b wfapi.ResourceBuilder)) {
	resourceChild(b, bld)
}

func (b *iteratorBuilder) Workflow(bld func(b wfapi.WorkflowBuilder)) {
	workflowChild(b, bld)
}

func (b *iteratorBuilder) Stateless(bld func(b wfapi.StatelessBuilder)) {
	statelessChild(b, bld)
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

func (b *iteratorBuilder) Build() wfapi.Activity {
	b.validate()
	return NewIterator(b.GetName(), b.when, b.input, b.output, b.style, b.children[0], b.over, b.variables)
}

func (b *iteratorBuilder) validate() {
	if len(b.children) != 1 {
		panic(eval.Error(wfapi.WF_ITERATOR_NOT_ONE_ACTIVITY, issue.NO_ARGS))
	}
}

type resourceBuilder struct {
	builder
	state wfapi.State
}

func (b *resourceBuilder) Build() wfapi.Activity {
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

func (b *workflowBuilder) Build() wfapi.Activity {
	b.validate()
	return NewWorkflow(b.GetName(), b.when, b.input, b.output, b.children)
}

func (b *workflowBuilder) Action(bld func(b wfapi.ActionBuilder)) {
	actionChild(b, bld)
}

func (b *workflowBuilder) Resource(bld func(b wfapi.ResourceBuilder)) {
	resourceChild(b, bld)
}

func (b *workflowBuilder) Workflow(bld func(b wfapi.WorkflowBuilder)) {
	workflowChild(b, bld)
}

func (b *workflowBuilder) Stateless(bld func(b wfapi.StatelessBuilder)) {
	statelessChild(b, bld)
}

func (b *workflowBuilder) Iterator(bld func(b wfapi.IteratorBuilder)) {
	ab := &iteratorBuilder{childBuilder: childBuilder{builder: builder{parent: b, ctx: b.ctx, when: condition.Always, input: eval.NoParameters, output: eval.NoParameters}}}
	bld(ab)
	b.AddChild(ab)
}

type statelessBuilder struct {
	builder
	function interface{}
}

func (b *statelessBuilder) Build() wfapi.Activity {
	b.validate()
	return NewStateless(b.GetName(), b.when, b.input, b.output, b.function)
}

func (b *statelessBuilder) Doer(d interface{}) {
	b.function = d
}
