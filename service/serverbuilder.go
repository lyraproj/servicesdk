package service

import (
	"reflect"
	"sort"
	"strings"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/condition"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wfapi"
)

type GoState struct {
	t eval.ObjectType
	v reflect.Value
}

func NewGoState(t eval.ObjectType, v reflect.Value) *GoState {
	return &GoState{t, v}
}

func (s *GoState) Type() eval.ObjectType {
	return s.t
}

func (s *GoState) State() interface{} {
	return s.v
}

func GoStateConverter(c eval.Context, state wfapi.State, input eval.OrderedMap) eval.PuppetObject {
	return eval.WrapReflected(c, state.State().(reflect.Value)).(eval.PuppetObject)
}

type ServerBuilder struct {
	ctx             eval.Context
	serviceId       eval.TypedName
	stateConv       wfapi.StateConverter
	types           map[string]eval.Type
	handlerFor      map[string]eval.Type
	activities      map[string]serviceapi.Definition
	callables       map[string]reflect.Value
	states          map[string]wfapi.State
	callableObjects []eval.PuppetObject
}

func NewServerBuilder(ctx eval.Context, serviceName string) *ServerBuilder {
	return &ServerBuilder{
		ctx:             ctx,
		serviceId:       eval.NewTypedName(eval.NsService, assertTypeName(serviceName)),
		callables:       make(map[string]reflect.Value),
		callableObjects: make([]eval.PuppetObject, 0),
		handlerFor:      make(map[string]eval.Type),
		activities:      make(map[string]serviceapi.Definition),
		types:           make(map[string]eval.Type),
		states:          make(map[string]wfapi.State)}
}

func assertTypeName(name string) string {
	if types.QREF_PATTERN.MatchString(name) {
		return name
	}
	panic(eval.Error(WF_ILLEGAL_TYPE_NAME, issue.H{`name`: name}))
}

func (ds *ServerBuilder) RegisterStateConverter(sf wfapi.StateConverter) {
	ds.stateConv = sf
}

// RegisterAPI registers a struct as an invokable. The callable instance given as the argument becomes the
// actual receiver the calls.
func (ds *ServerBuilder) RegisterAPI(name string, callable interface{}) {
	name = assertTypeName(name)
	if po, ok := callable.(eval.PuppetObject); ok {
		ds.callableObjects = append(ds.callableObjects, po)
	} else {
		rv := reflect.ValueOf(callable)
		rt := rv.Type()
		pt, ok := ds.ctx.ImplementationRegistry().ReflectedToType(rt)
		if !ok {
			pt = ds.ctx.Reflector().TypeFromReflect(name, nil, rt)
		}
		if _, ok := ds.types[name]; !ok {
			ds.registerType(name, pt)
		}
		ds.registerCallable(name, rv)
	}
}

// RegisterAPIType registers a the type of a struct as an invokable type. The struct should be a zero
// value. This method must be used to ensure that all type info is present for callable instances added to an
// already created service
func (ds *ServerBuilder) RegisterApiType(name string, callable interface{}) {
	name = assertTypeName(name)
	rv := reflect.ValueOf(callable)
	rt := rv.Type()
	pt, ok := ds.ctx.ImplementationRegistry().ReflectedToType(rt)
	if !ok {
		pt = ds.ctx.Reflector().TypeFromReflect(name, nil, rt)
	}
	if _, ok := ds.types[name]; !ok {
		ds.registerType(name, pt)
	}
}

// RegisterState registers the unresolved state of a resource.
func (ds *ServerBuilder) RegisterState(name string, state wfapi.State) {
	t := state.Type()
	if _, ok := ds.types[t.Name()]; !ok {
		ds.RegisterType(t)
	}
	ds.states[name] = state
}

func (ds *ServerBuilder) BuildResource(goType interface{}, bld func(f ResourceTypeBuilder)) eval.AnnotatedType {
	rb := &rtBuilder{ctx: ds.ctx}
	bld(rb)
	return rb.Build(goType)
}

// RegisterHandler registers a callable struct as an invokable capable of handling a state described using
// eval.Type. The callable instance given as the argument becomes the actual receiver the calls.
func (ds *ServerBuilder) RegisterHandler(name string, callable interface{}, stateType eval.Type) {
	ds.RegisterAPI(name, callable)
	ds.types[stateType.Name()] = stateType
	ds.handlerFor[name] = stateType
}

// RegisterTypes registers arbitrary Go types to the TypeSet exported by this service.
//
// A value is typically a pointer to the zero value of a struct. The name of the generated type for
// that struct will be the struct name prefixed by the service ID.
func (ds *ServerBuilder) RegisterTypes(namespace string, values ...interface{}) []eval.Type {
	ts := make([]eval.Type, len(values))
	for i, v := range values {
		switch v.(type) {
		case eval.Type:
			t := v.(eval.Type)
			ds.types[t.Name()] = t
			ts[i] = t
		case eval.AnnotatedType:
			ts[i] = ds.registerReflectedType(namespace, v.(eval.AnnotatedType))
		case reflect.Type:
			ts[i] = ds.registerReflectedType(namespace, eval.NewTaggedType(v.(reflect.Type), nil))
		case reflect.Value:
			ts[i] = ds.registerReflectedType(namespace, eval.NewTaggedType(v.(reflect.Value).Type(), nil))
		default:
			ts[i] = ds.registerReflectedType(namespace, eval.NewTaggedType(reflect.TypeOf(v), nil))
		}
	}
	return ts
}

func (ds *ServerBuilder) registerReflectedType(namespace string, tg eval.AnnotatedType) eval.Type {
	typ := tg.Type()
	if typ.Kind() == reflect.Ptr {
		el := typ.Elem()
		if el.Kind() != reflect.Interface {
			typ = el
		}
	}

	parent := types.ParentType(typ)
	var pt eval.Type
	if parent != nil {
		pt = ds.registerReflectedType(namespace, eval.NewTaggedType(parent, nil))
	}

	name := namespace + `::` + typ.Name()
	et, ok := ds.types[name]
	if ok {
		// Type is already registered
		return et
	}

	et = ds.ctx.Reflector().TypeFromTagged(name, pt, tg)
	ds.types[name] = et
	return et
}

// RegisterActivity registers an activity
func (ds *ServerBuilder) RegisterActivity(activity wfapi.Activity) {
	name := activity.Name()
	if _, found := ds.activities[name]; found {
		panic(eval.Error(WF_ALREADY_REGISTERED, issue.H{`namespace`: eval.NsDefinition, `identifier`: name}))
	}
	ds.activities[name] = ds.createActivityDefinition(activity)
}

func (ds *ServerBuilder) registerCallable(name string, callable reflect.Value) {
	if _, found := ds.callables[name]; found {
		panic(eval.Error(WF_ALREADY_REGISTERED, issue.H{`namespace`: eval.NsInterface, `identifier`: name}))
	}
	ds.callables[name] = callable
}

func (ds *ServerBuilder) RegisterType(typ eval.Type) {
	ds.registerType(typ.Name(), typ)
}

func (ds *ServerBuilder) registerType(name string, typ eval.Type) {
	if _, found := ds.types[name]; found {
		panic(eval.Error(WF_ALREADY_REGISTERED, issue.H{`namespace`: eval.NsType, `identifier`: name}))
	}
	ds.types[name] = typ
}

func (ds *ServerBuilder) createActivityDefinition(activity wfapi.Activity) serviceapi.Definition {
	props := make([]*types.HashEntry, 0, 5)

	if input := paramsAsList(activity.Input()); input != nil {
		props = append(props, types.WrapHashEntry2(`input`, input))
	}
	if output := paramsAsList(activity.Output()); output != nil {
		props = append(props, types.WrapHashEntry2(`output`, output))
	}
	if activity.When() != condition.Always {
		props = append(props, types.WrapHashEntry2(`when`, types.WrapString(activity.When().String())))
	}

	name := activity.Name()
	var style string
	switch activity.(type) {
	case wfapi.Workflow:
		style = `workflow`
		props = append(props, types.WrapHashEntry2(`activities`, ds.activitiesAsList(activity.(wfapi.Workflow).Activities())))
	case wfapi.Resource:
		style = `resource`
		state := activity.(wfapi.Resource).State()
		ds.RegisterState(name, state)
		props = append(props, types.WrapHashEntry2(`resource_type`, state.Type()))
	case wfapi.Action:
		style = `action`
		ds.RegisterAPI(name, activity.(wfapi.Action).Interface())
		props = append(props, types.WrapHashEntry2(`interface`, ds.types[name]))
	case wfapi.Stateless:
		style = `stateless`
		ds.RegisterAPI(name, activity.(wfapi.Stateless).Function())
		props = append(props, types.WrapHashEntry2(`interface`, ds.types[name]))
	case wfapi.Iterator:
		style = `iterator`
		iter := activity.(wfapi.Iterator)
		props = append(props, types.WrapHashEntry2(`iteration_style`, types.WrapString(iter.IterationStyle().String())))
		props = append(props, types.WrapHashEntry2(`over`, paramsAsList(iter.Over())))
		props = append(props, types.WrapHashEntry2(`variables`, paramsAsList(iter.Variables())))
		props = append(props, types.WrapHashEntry2(`producer`, ds.createActivityDefinition(iter.Producer())))
	}
	props = append(props, types.WrapHashEntry2(`style`, types.WrapString(style)))
	return serviceapi.NewDefinition(eval.NewTypedName(eval.NsDefinition, name), ds.serviceId, types.WrapHash(props))
}

func paramsAsList(params []eval.Parameter) eval.List {
	np := len(params)
	if np == 0 {
		return nil
	}
	ps := make([]eval.Value, np)
	for i, p := range params {
		ps[i] = p
	}
	return types.WrapValues(ps)
}

func (ds *ServerBuilder) activitiesAsList(activities []wfapi.Activity) eval.List {
	as := make([]eval.Value, len(activities))
	for i, a := range activities {
		as[i] = ds.createActivityDefinition(a)
	}
	return types.WrapValues(as)
}

func CreateTypeSet(ts map[string]eval.Type) eval.TypeSet {
	result := make(map[string]interface{})
	for k, t := range ts {
		addName(strings.Split(k, `::`), result, t)
	}

	if len(result) != 1 {
		panic(eval.Error(WF_NO_COMMON_NAMESPACE, issue.NO_ARGS))
	}

next:
	for {
		// If the value below is a map of size 1, then move that map up
		for k, v := range result {
			if sm, ok := v.(map[string]interface{}); ok && len(sm) == 1 {
				delete(result, k)
				for sk, sv := range sm {
					result[k+`::`+sk] = sv
				}
				continue next
			}
			break next
		}
	}
	t := makeType(``, result)
	if ts, ok := t.(eval.TypeSet); ok {
		return ts
	}

	sgs := strings.Split(t.Name(), `::`)
	tsn := strings.Join(sgs[:len(sgs)-1], `::`)
	tn := sgs[len(sgs)-1]
	es := make([]*types.HashEntry, 0)
	es = append(es, types.WrapHashEntry2(eval.KEY_PCORE_URI, types.WrapString(string(eval.PCORE_URI))))
	es = append(es, types.WrapHashEntry2(eval.KEY_PCORE_VERSION, types.WrapSemVer(eval.PCORE_VERSION)))
	es = append(es, types.WrapHashEntry2(types.KEY_VERSION, types.WrapSemVer(ServerVersion)))
	es = append(es, types.WrapHashEntry2(types.KEY_TYPES, types.SingletonHash2(tn, t)))
	return types.NewTypeSetType(eval.RUNTIME_NAME_AUTHORITY, tsn, types.WrapHash(es))
}

func makeType(name string, tree map[string]interface{}) eval.Type {
	rl := len(tree)
	ts := make(map[string]eval.Type, rl)
	for k, v := range tree {
		var t eval.Type
		if x, ok := v.(eval.Type); ok {
			t = x
		} else {
			var tn string
			if name == `` {
				tn = k
			} else {
				tn = name + `::` + k
			}
			t = makeType(tn, v.(map[string]interface{}))
		}
		if rl == 1 {
			return t
		}
		ts[k] = t
	}
	es := make([]*types.HashEntry, 0)
	es = append(es, types.WrapHashEntry2(eval.KEY_PCORE_URI, types.WrapString(string(eval.PCORE_URI))))
	es = append(es, types.WrapHashEntry2(eval.KEY_PCORE_VERSION, types.WrapSemVer(eval.PCORE_VERSION)))
	es = append(es, types.WrapHashEntry2(types.KEY_VERSION, types.WrapSemVer(ServerVersion)))
	es = append(es, types.WrapHashEntry2(types.KEY_TYPES, types.WrapStringToTypeMap(ts)))
	return types.NewTypeSetType(eval.RUNTIME_NAME_AUTHORITY, name, types.WrapHash(es))
}

func addName(ks []string, tree map[string]interface{}, t eval.Type) {
	kl := len(ks)
	k0 := ks[0]
	if sn, ok := tree[k0]; ok {
		if sm, ok := sn.(map[string]interface{}); ok {
			if kl > 1 {
				addName(ks[1:], sm, t)
				return
			}
		}
		panic(`type/typeset clash`)
	}
	if kl > 1 {
		sm := make(map[string]interface{})
		tree[k0] = sm
		addName(ks[1:], sm, t)
	} else {
		tree[k0] = t
	}
}

func (ds *ServerBuilder) Server() *Server {
	ts := CreateTypeSet(ds.types)
	ds.ctx.AddTypes(ts)

	defs := make([]eval.Value, 0, len(ds.callables)+len(ds.activities))

	callableStyle := types.WrapString(`callable`)
	// Create invokable definitions for callables
	for k := range ds.callables {
		props := make([]*types.HashEntry, 0, 2)
		props = append(props, types.WrapHashEntry2(`interface`, ds.types[k]))
		props = append(props, types.WrapHashEntry2(`style`, callableStyle))
		if stateType, ok := ds.handlerFor[k]; ok {
			props = append(props, types.WrapHashEntry2(`handler_for`, stateType))
		}
		defs = append(defs, serviceapi.NewDefinition(eval.NewTypedName(eval.NsDefinition, k), ds.serviceId, types.WrapHash(props)))
	}

	for _, po := range ds.callableObjects {
		k := po.PType().Name()
		props := make([]*types.HashEntry, 0, 2)
		props = append(props, types.WrapHashEntry2(`interface`, po.PType()))
		props = append(props, types.WrapHashEntry2(`style`, callableStyle))
		if stateType, ok := ds.handlerFor[k]; ok {
			props = append(props, types.WrapHashEntry2(`handler_for`, stateType))
		}
		defs = append(defs, serviceapi.NewDefinition(eval.NewTypedName(eval.NsDefinition, k), ds.serviceId, types.WrapHash(props)))
	}

	// Add registered activities
	for _, a := range ds.activities {
		defs = append(defs, a)
	}
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].(serviceapi.Definition).Identifier().Name() < defs[j].(serviceapi.Definition).Identifier().Name()
	})

	callables := make(map[string]eval.Value, len(ds.callables)+len(ds.callableObjects))
	for k, v := range ds.callables {
		callables[k] = eval.WrapReflected(ds.ctx, v)
	}

	for _, po := range ds.callableObjects {
		callables[po.PType().Name()] = po
	}

	return &Server{context: ds.ctx, id: ds.serviceId, typeSet: ts, metadata: types.WrapValues(defs), stateConv: ds.stateConv, callables: callables, states: ds.states}
}
