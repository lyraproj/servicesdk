package wf

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-servicesdk/condition"
	"github.com/puppetlabs/go-servicesdk/wfapi"
)

func init() {
	wfapi.NewWorkflow = func(ctx eval.Context, bf func(wfapi.WorkflowBuilder)) wfapi.Workflow {
		bld := &workflowBuilder{builder: builder{ctx: ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
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
}

func (b *builder) Name(n string) {
	b.name = n
}

func (b *builder) When(w string) {
	b.when = condition.Parse(w)
}

func (b *builder) Input(name, typeName string) {
	param := impl.NewParameter(name, b.ctx.ParseType2(typeName), nil, false)
	if len(b.input) == 0 {
		b.input = []eval.Parameter{param}
	} else {
		b.input = append(b.input, param)
	}
}

func (b *builder) Output(name, typeName string) {
	param := impl.NewParameter(name, b.ctx.ParseType2(typeName), nil, false)
	if len(b.output) == 0 {
		b.output = []eval.Parameter{param}
	} else {
		b.output = append(b.output, param)
	}
}

type actionBuilder struct {
	builder
	crd wfapi.CRD
}

func (b *actionBuilder) CRD(c wfapi.CRD) {
	b.crd = c
}

func (b *actionBuilder) build() wfapi.Action {
	return NewAction(b.name, b.when, b.input, b.output, b.crd)
}

type resourceBuilder struct {
	builder
	state eval.PuppetObject
}

func (b *resourceBuilder) build() wfapi.Resource {
	return NewResource(b.name, b.when, b.input, b.output, b.state)
}

func (b *resourceBuilder) State(o interface{}) {
	b.state = eval.Wrap(b.ctx, o).(eval.PuppetObject)
}

type workflowBuilder struct {
	builder
	activities []wfapi.Activity
}

func (b *workflowBuilder) build() wfapi.Workflow {
	return NewWorkflow(b.name, b.when, b.input, b.output, b.activities)
}

func (b *workflowBuilder) Action(bld func(b wfapi.ActionBuilder)) {
	ab := &actionBuilder{builder: builder{ctx: b.ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
	bld(ab)
	b.activities = append(b.activities, ab.build())
}

func (b *workflowBuilder) Resource(bld func(b wfapi.ResourceBuilder)) {
	ab := &resourceBuilder{builder: builder{ctx: b.ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
	bld(ab)
	b.activities = append(b.activities, ab.build())
}

func (b *workflowBuilder) Workflow(bld func(b wfapi.WorkflowBuilder)) {
	ab := &workflowBuilder{builder: builder{ctx: b.ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
	bld(ab)
	b.activities = append(b.activities, ab.build())
}

func (b *workflowBuilder) Stateless(bld func(b wfapi.StatelessBuilder)) {
	ab := &statelessBuilder{builder: builder{ctx: b.ctx, when: condition.Always, input:eval.NoParameters, output:eval.NoParameters}}
	bld(ab)
	b.activities = append(b.activities, ab.build())
}

type statelessBuilder struct {
	builder
	doer wfapi.Doer
}

func (b *statelessBuilder) build() wfapi.Stateless {
	return NewStateless(b.name, b.when, b.input, b.output, b.doer)
}

func (b *statelessBuilder) Doer(d wfapi.Doer) {
	b.doer = d
}
