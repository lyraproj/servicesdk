package service

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-servicesdk/condition"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/puppetlabs/go-servicesdk/wfapi"
	"reflect"
	"sort"
)

type ServerBuilder struct {
	ctx        eval.Context
	serviceId  string
	types      map[string]eval.Type
	handlerFor map[string]reflect.Value
	activities map[string]serviceapi.Definition
	callables  map[string]reflect.Value
	callableObjects []eval.PuppetObject
}

func NewServerBuilder(ctx eval.Context, serviceName string) *ServerBuilder {
	return &ServerBuilder{
		ctx:        ctx,
		serviceId:  assertTypeName(serviceName),
		callables:  make(map[string]reflect.Value),
		callableObjects: make([]eval.PuppetObject, 0),
		handlerFor: make(map[string]reflect.Value),
		activities: make(map[string]serviceapi.Definition),
		types:      make(map[string]eval.Type)}
}

func assertTypeName(name string) string {
	if types.QREF_PATTERN.MatchString(name) {
		return name
	}
	panic(eval.Error(WF_ILLEGAL_TYPE_NAME, issue.H{`name`: name}))
}

// RegisterAPI registers a struct as an invokable. The callable instance given as the argument becomes the
// actual receiver the calls.
func (ds *ServerBuilder) RegisterAPI(name string, callable interface{}) {
	name = assertTypeName(name)
	if po, ok := callable.(eval.PuppetObject); ok {
		ds.callableObjects = append(ds.callableObjects, po)
	} else {
		rv := reflect.ValueOf(callable)
		ds.registerType(name, ds.ctx.Reflector().ObjectTypeFromReflect(ds.serviceId+`::`+name, nil, rv.Type()))
		ds.registerCallable(name, rv)
	}
}

// RegisterHandler registers a callable struct as an invokable capable of handling a state. The
// callable instance given as the argument becomes the actual receiver the calls.
func (ds *ServerBuilder) RegisterHandler(name string, callable interface{}, state interface{}) {
	ds.RegisterAPI(name, callable)
	ds.handlerFor[name] = reflect.ValueOf(state)
}

// RegisterTypes registers arbitrary Go types to the TypeSet exported by this service.
//
// A value is typically a pointer to the zero value of a struct. The name of the generated type for
// that struct will be the struct name prefixed by the service ID.
func (ds *ServerBuilder) RegisterTypes(values ...interface{}) {
	for _, v := range values {
		ds.registerReflectedType(reflect.TypeOf(v))
	}
}

func (ds *ServerBuilder) registerReflectedType(typ reflect.Type) eval.Type {
	if typ.Kind() == reflect.Ptr {
		el := typ.Elem()
		if el.Kind() != reflect.Interface {
			typ = el
		}
	}

	parent := types.ParentType(typ)
	var pt eval.Type
	if parent != nil {
		pt = ds.registerReflectedType(parent)
	}

	name := typ.Name()
	et, ok := ds.types[name]
	if ok {
		// Type is already registered
		return et
	}

	et = ds.ctx.Reflector().ObjectTypeFromReflect(ds.serviceId+`::`+name, pt, typ)
	ds.types[name] = et
	return et
}

// RegisterActivity registers an activity
func (ds *ServerBuilder) RegisterActivity(activity wfapi.Activity) {
	name := activity.Name()
	if _, found := ds.activities[name]; found {
		panic(eval.Error(WF_ALREADY_REGISTERED, issue.H{`namespace`: eval.ACTIVITY, `identifier`: name}))
	}
	ds.activities[name] = ds.createActivityDefinition(eval.NewTypedName(serviceapi.NsService, ds.serviceId), activity)
}

func (ds *ServerBuilder) registerCallable(name string, callable reflect.Value) {
	if _, found := ds.callables[name]; found {
		panic(eval.Error(WF_ALREADY_REGISTERED, issue.H{`namespace`: serviceapi.NsInterface, `identifier`: name}))
	}
	ds.callables[name] = callable
}

func (ds *ServerBuilder) registerType(name string, typ eval.Type) {
	if _, found := ds.types[name]; found {
		panic(eval.Error(WF_ALREADY_REGISTERED, issue.H{`namespace`: eval.TYPE, `identifier`: name}))
	}
	ds.types[name] = typ
}

func (ds *ServerBuilder) createActivityDefinition(serviceId eval.TypedName, activity wfapi.Activity) serviceapi.Definition {
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
		props = append(props, types.WrapHashEntry2(`activities`, ds.activitiesAsList(serviceId, activity.(wfapi.Workflow).Activities())))
	case wfapi.Resource:
		style = `resource`
		sb := activity.(wfapi.Resource).State()
		retrieverName := `Get` + name
		ds.RegisterAPI(retrieverName, sb)
		props = append(props, types.WrapHashEntry2(`state`, types.WrapString(ds.types[retrieverName].Name())))
	case wfapi.Action:
		style = `action`
		ds.RegisterAPI(name, activity.(wfapi.Action).Interface())
		props = append(props, types.WrapHashEntry2(`interface`, ds.types[name]))
	case wfapi.Stateless:
		style = `stateless`
		fc := activity.(wfapi.Stateless).Interface()
		ds.RegisterAPI(name, fc)
		props = append(props, types.WrapHashEntry2(`interface`, ds.types[name]))
	}
	props = append(props, types.WrapHashEntry2(`style`, types.WrapString(style)))
	return serviceapi.NewDefinition(eval.NewTypedName(serviceapi.NsActivity, name), serviceId, types.WrapHash(props))
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

func (ds *ServerBuilder) activitiesAsList(serviceId eval.TypedName, activities []wfapi.Activity) eval.List {
	as := make([]eval.Value, len(activities))
	for i, a := range activities {
		as[i] = ds.createActivityDefinition(serviceId, a)
	}
	return types.WrapValues(as)
}

func (ds *ServerBuilder) Server() *Server {
	es := make([]*types.HashEntry, 0)
	es = append(es, types.WrapHashEntry2(eval.KEY_PCORE_URI, types.WrapString(string(eval.PCORE_URI))))
	es = append(es, types.WrapHashEntry2(eval.KEY_PCORE_VERSION, types.WrapSemVer(eval.PCORE_VERSION)))
	es = append(es, types.WrapHashEntry2(types.KEY_VERSION, types.WrapSemVer(ServerVersion)))
	es = append(es, types.WrapHashEntry2(types.KEY_TYPES, types.WrapStringToTypeMap(ds.types)))
	ts := types.NewTypeSetType(eval.RUNTIME_NAME_AUTHORITY, ds.serviceId, types.WrapHash(es))
	ds.ctx.AddTypes(ts)

	serviceId := eval.NewTypedName(serviceapi.NsService, ds.serviceId)
	defs := make([]eval.Value, 0, len(ds.callables)+len(ds.activities))

	// Create invokable definitions for callables
	for k := range ds.callables {
		if state, ok := ds.handlerFor[k]; ok {
			props := make([]*types.HashEntry, 0, 2)
			props = append(props, types.WrapHashEntry2(`interface`, types.WrapString(ds.types[k].Name())))
			props = append(props, types.WrapHashEntry2(`handlerFor`, eval.WrapReflected(ds.ctx, state).PType()))
			defs = append(defs, serviceapi.NewDefinition(eval.NewTypedName(serviceapi.NsActivity, k), serviceId, types.WrapHash(props)))
		}
	}

	for _, po := range ds.callableObjects {
		k := po.(issue.Named).Name()
		if state, ok := ds.handlerFor[k]; ok {
			props := make([]*types.HashEntry, 0, 2)
			props = append(props, types.WrapHashEntry2(`interface`, types.WrapString(po.PType().Name())))
			props = append(props, types.WrapHashEntry2(`handlerFor`, eval.WrapReflected(ds.ctx, state).PType()))
			defs = append(defs, serviceapi.NewDefinition(eval.NewTypedName(serviceapi.NsActivity, k), serviceId, types.WrapHash(props)))
		}
	}

	// Add registered activities
	for _, a := range ds.activities {
		defs = append(defs, a)
	}
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].(serviceapi.Definition).Identifier().Name() < defs[j].(serviceapi.Definition).Identifier().Name()
	})

	callables := make(map[string]eval.Value, len(ds.callables) + len(ds.callableObjects))
	for k, v := range ds.callables {
		callables[k] = eval.WrapReflected(ds.ctx, v)
	}

	for _, po := range ds.callableObjects {
		callables[po.(issue.Named).Name()] = po
	}

	return &Server{context: ds.ctx, typeSet: ts, metadata: types.WrapValues(defs), callables: callables}
}
