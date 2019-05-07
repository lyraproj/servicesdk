package service

import (
	"reflect"
	"sort"
	"strings"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wf"
)

type Builder struct {
	ctx             px.Context
	serviceId       px.TypedName
	stateConverter  wf.StateConverter
	types           map[string]px.Type
	handlerFor      map[string]px.Type
	steps           map[string]serviceapi.Definition
	callables       map[string]reflect.Value
	states          map[string]wf.State
	callableObjects map[string]px.PuppetObject
}

func NewServiceBuilder(ctx px.Context, serviceName string) *Builder {
	return &Builder{
		ctx:             ctx,
		serviceId:       px.NewTypedName(px.NsService, assertTypeName(serviceName)),
		callables:       make(map[string]reflect.Value),
		callableObjects: make(map[string]px.PuppetObject),
		handlerFor:      make(map[string]px.Type),
		steps:           make(map[string]serviceapi.Definition),
		types:           make(map[string]px.Type),
		states:          make(map[string]wf.State)}
}

func assertTypeName(name string) string {
	if types.TypeNamePattern.MatchString(name) {
		return name
	}
	panic(px.Error(IllegalTypeName, issue.H{`name`: name}))
}

func (ds *Builder) RegisterStateConverter(sf wf.StateConverter) {
	ds.stateConverter = sf
}

// RegisterAPI registers a struct as an invokable. The callable instance given as the argument becomes the
// actual receiver the calls.
func (ds *Builder) RegisterAPI(name string, callable interface{}) {
	name = assertTypeName(name)
	if po, ok := callable.(px.PuppetObject); ok {
		ds.callableObjects[name] = po
	} else {
		rv := reflect.ValueOf(callable)
		rt := rv.Type()
		_, ok := ds.ctx.ImplementationRegistry().ReflectedToType(rt)
		if !ok {
			pt := ds.ctx.Reflector().TypeFromReflect(name, nil, rt)
			ds.registerType(name, pt)
		}
		ds.registerCallable(name, rv)
	}
}

// RegisterAPIType registers a the type of a struct as an invokable type. The struct should be a zero
// value. This method must be used to ensure that all type info is present for callable instances added to an
// already created service
func (ds *Builder) RegisterApiType(name string, callable interface{}) {
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
func (ds *Builder) RegisterState(name string, state wf.State) {
	ds.states[name] = state
}

func (ds *Builder) BuildResource(goType interface{}, bld func(f ResourceTypeBuilder)) px.AnnotatedType {
	rb := &rtBuilder{ctx: ds.ctx}
	bld(rb)
	return rb.Build(goType)
}

// RegisterHandler registers a callable struct as an invokable capable of handling a state described using
// px.Type. The callable instance given as the argument becomes the actual receiver the calls.
func (ds *Builder) RegisterHandler(name string, callable interface{}, stateType px.Type) {
	ds.RegisterAPI(name, callable)
	ds.types[stateType.Name()] = stateType
	ds.handlerFor[name] = stateType
}

// RegisterTypes registers arbitrary Go types to the TypeSet exported by this service.
//
// A value is typically a pointer to the zero value of a struct. The name of the generated type for
// that struct will be the struct name prefixed by the service ID.
func (ds *Builder) RegisterTypes(namespace string, values ...interface{}) []px.Type {
	ts := make([]px.Type, len(values))
	for i, v := range values {
		switch v := v.(type) {
		case px.Type:
			ds.types[v.Name()] = v
			ts[i] = v
		case px.AnnotatedType:
			ts[i] = ds.registerReflectedType(namespace, v)
		case reflect.Type:
			ts[i] = ds.registerReflectedType(namespace, px.NewTaggedType(v, nil))
		case reflect.Value:
			ts[i] = ds.registerReflectedType(namespace, px.NewTaggedType(v.Type(), nil))
		default:
			ts[i] = ds.registerReflectedType(namespace, px.NewTaggedType(reflect.TypeOf(v), nil))
		}
	}
	return ts
}

func (ds *Builder) registerReflectedType(namespace string, tg px.AnnotatedType) px.Type {
	typ := tg.Type()
	if typ.Kind() == reflect.Ptr {
		el := typ.Elem()
		if el.Kind() != reflect.Interface {
			typ = el
		}
	}

	parent := types.ParentType(typ)
	var pt px.Type
	if parent != nil {
		pt = ds.registerReflectedType(namespace, px.NewTaggedType(parent, nil))
	}

	name := namespace + `::` + typ.Name()
	et, ok := ds.types[name]
	if ok {
		// Type is already registered
		return et
	}

	var registerFieldType func(ft reflect.Type)
	registerFieldType = func(ft reflect.Type) {
		switch ft.Kind() {
		case reflect.Slice, reflect.Ptr, reflect.Array, reflect.Map:
			registerFieldType(ft.Elem())
		case reflect.Struct:
			if ft == parent {
				break
			}
			// Register type unless it's already registered
			if _, err := px.WrapReflectedType(ds.ctx, ft); err != nil {
				ds.registerReflectedType(namespace, px.NewAnnotatedType(ft, nil, nil))
			}
		}
	}

	et = ds.ctx.Reflector().TypeFromTagged(name, pt, tg, func() {
		// Register nested types unless already known to the implementation registry
		nf := typ.NumField()
		for i := 0; i < nf; i++ {
			f := typ.Field(i)
			if f.PkgPath == `` {
				// Exported
				registerFieldType(f.Type)
			}
		}
	})
	ds.types[name] = et
	return et
}

// RegisterStep registers an step
func (ds *Builder) RegisterStep(step wf.Step) {
	name := step.Name()
	if _, found := ds.steps[name]; found {
		panic(px.Error(AlreadyRegistered, issue.H{`namespace`: px.NsDefinition, `identifier`: name}))
	}
	ds.steps[name] = ds.createStepDefinition(step)
}

func (ds *Builder) registerCallable(name string, callable reflect.Value) {
	if _, found := ds.callables[name]; found {
		panic(px.Error(AlreadyRegistered, issue.H{`namespace`: px.NsInterface, `identifier`: name}))
	}
	ds.callables[name] = callable
}

func (ds *Builder) RegisterType(typ px.Type) {
	ds.registerType(typ.Name(), typ)
}

func (ds *Builder) registerType(name string, typ px.Type) {
	if _, found := ds.types[name]; found {
		panic(px.Error(AlreadyRegistered, issue.H{`namespace`: px.NsType, `identifier`: name}))
	}
	ds.types[name] = typ
}

func (ds *Builder) createStepDefinition(step wf.Step) serviceapi.Definition {
	props := make([]*types.HashEntry, 0, 5)

	if parameters := paramsAsList(step.Parameters()); parameters != nil {
		props = append(props, types.WrapHashEntry2(`parameters`, parameters))
	}
	if returns := paramsAsList(step.Returns()); returns != nil {
		props = append(props, types.WrapHashEntry2(`returns`, returns))
	}
	if step.When() != wf.Always {
		props = append(props, types.WrapHashEntry2(`when`, types.WrapString(step.When().String())))
	}

	name := step.Name()
	var style string
	switch step := step.(type) {
	case wf.Workflow:
		style = `workflow`
		props = append(props, types.WrapHashEntry2(`steps`, ds.stepsAsList(step.Steps())))
	case wf.Resource:
		style = `resource`
		state := step.State()
		extId := step.ExternalId()
		ds.RegisterState(name, state)
		props = append(props, types.WrapHashEntry2(`resourceType`, state.Type()))
		if extId != `` {
			props = append(props, types.WrapHashEntry2(`externalId`, types.WrapString(extId)))
		}
	case wf.StateHandler:
		style = `stateHandler`
		tn := strings.Title(name)
		api := step.Interface()
		ds.RegisterAPI(tn, api)
		var ifd px.Type
		if po, ok := api.(px.PuppetObject); ok {
			ifd = po.PType()
		} else {
			ifd = ds.types[tn]
		}
		props = append(props, types.WrapHashEntry2(`interface`, ifd))
	case wf.Action:
		style = `action`
		tn := strings.Title(name)
		api := step.Function()
		ds.RegisterAPI(tn, api)
		var ifd px.Type
		if po, ok := api.(px.PuppetObject); ok {
			ifd = po.PType()
		} else {
			ifd, ok = ds.types[tn]
			if !ok {
				ifd, _ = ds.ctx.ImplementationRegistry().ReflectedToType(reflect.TypeOf(api))
			}
		}
		props = append(props, types.WrapHashEntry2(`interface`, ifd))
	case wf.Iterator:
		style = `iterator`
		props = append(props, types.WrapHashEntry2(`iterationStyle`, types.WrapString(step.IterationStyle().String())))
		props = append(props, types.WrapHashEntry2(`over`, step.Over()))
		vars := step.Variables()
		if len(vars) > 0 {
			props = append(props, types.WrapHashEntry2(`variables`, paramsAsList(vars)))
		}
		if step.Into() != `` {
			props = append(props, types.WrapHashEntry2(`into`, types.WrapString(step.Into())))
		}
		props = append(props, types.WrapHashEntry2(`producer`, ds.createStepDefinition(step.Producer())))
	case wf.Reference:
		style = `reference`
		props = append(props, types.WrapHashEntry2(`reference`, types.WrapString(step.Reference())))
	}
	props = append(props, types.WrapHashEntry2(`style`, types.WrapString(style)))
	return serviceapi.NewDefinition(px.NewTypedName(px.NsDefinition, name), ds.serviceId, types.WrapHash(props))
}

func paramsAsList(params []px.Parameter) px.List {
	np := len(params)
	if np == 0 {
		return nil
	}
	ps := make([]px.Value, np)
	for i, p := range params {
		ps[i] = p
	}
	return types.WrapValues(ps)
}

func (ds *Builder) stepsAsList(steps []wf.Step) px.List {
	as := make([]px.Value, len(steps))
	for i, a := range steps {
		as[i] = ds.createStepDefinition(a)
	}
	return types.WrapValues(as)
}

func CreateTypeSet(ts map[string]px.Type) px.TypeSet {
	result := make(map[string]interface{})
	for k, t := range ts {
		addName(strings.Split(k, `::`), result, t)
	}

	if len(result) != 1 {
		panic(px.Error(NoCommonNamespace, issue.NoArgs))
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
	if ts, ok := t.(px.TypeSet); ok {
		return ts
	}

	sgs := strings.Split(t.Name(), `::`)
	tsn := strings.Join(sgs[:len(sgs)-1], `::`)
	tn := sgs[len(sgs)-1]
	es := make([]*types.HashEntry, 0)
	es = append(es, types.WrapHashEntry2(px.KeyPcoreUri, types.WrapString(string(px.PcoreUri))))
	es = append(es, types.WrapHashEntry2(px.KeyPcoreVersion, types.WrapSemVer(px.PcoreVersion)))
	es = append(es, types.WrapHashEntry2(types.KeyVersion, types.WrapSemVer(ServerVersion)))
	es = append(es, types.WrapHashEntry2(types.KeyTypes, px.SingletonMap(tn, t)))
	return types.NewTypeSet(px.RuntimeNameAuthority, tsn, types.WrapHash(es))
}

func makeType(name string, tree map[string]interface{}) px.Type {
	rl := len(tree)
	ts := make(map[string]px.Type, rl)
	for k, v := range tree {
		var t px.Type
		if x, ok := v.(px.Type); ok {
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
	es = append(es, types.WrapHashEntry2(px.KeyPcoreUri, types.WrapString(string(px.PcoreUri))))
	es = append(es, types.WrapHashEntry2(px.KeyPcoreVersion, types.WrapSemVer(px.PcoreVersion)))
	es = append(es, types.WrapHashEntry2(types.KeyVersion, types.WrapSemVer(ServerVersion)))
	es = append(es, types.WrapHashEntry2(types.KeyTypes, types.WrapStringToTypeMap(ts)))
	return types.NewTypeSet(px.RuntimeNameAuthority, name, types.WrapHash(es))
}

func addName(ks []string, tree map[string]interface{}, t px.Type) {
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

func (ds *Builder) Server() *Server {
	var ts px.TypeSet
	if len(ds.types) > 0 {
		ts = CreateTypeSet(ds.types)
		px.AddTypes(ds.ctx, ts)
	}

	defs := make([]px.Value, 0, len(ds.callables)+len(ds.steps))

	callableStyle := types.WrapString(`callable`)
	// Create invokable definitions for callables
	for k, v := range ds.callables {
		props := make([]*types.HashEntry, 0, 2)
		if pt, ok := ds.ctx.ImplementationRegistry().ReflectedToType(v.Type()); ok {
			props = append(props, types.WrapHashEntry2(`interface`, pt))
		}

		props = append(props, types.WrapHashEntry2(`style`, callableStyle))
		if stateType, ok := ds.handlerFor[k]; ok {
			props = append(props, types.WrapHashEntry2(`handlerFor`, stateType))
		}
		defs = append(defs, serviceapi.NewDefinition(px.NewTypedName(px.NsDefinition, k+`::Api`), ds.serviceId, types.WrapHash(props)))
	}

	for k, po := range ds.callableObjects {
		props := make([]*types.HashEntry, 0, 2)
		props = append(props, types.WrapHashEntry2(`interface`, po.PType()))
		props = append(props, types.WrapHashEntry2(`style`, callableStyle))
		if stateType, ok := ds.handlerFor[k]; ok {
			props = append(props, types.WrapHashEntry2(`handlerFor`, stateType))
		}
		defs = append(defs, serviceapi.NewDefinition(px.NewTypedName(px.NsDefinition, k+`::Api`), ds.serviceId, types.WrapHash(props)))
	}

	// Add registered steps
	for _, a := range ds.steps {
		defs = append(defs, a)
	}
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].(serviceapi.Definition).Identifier().Name() < defs[j].(serviceapi.Definition).Identifier().Name()
	})

	callables := make(map[string]px.Value, len(ds.callables)+len(ds.callableObjects))
	for k, v := range ds.callables {
		callables[k] = px.WrapReflected(ds.ctx, v)
	}

	for k, po := range ds.callableObjects {
		callables[k] = po
	}

	return &Server{context: ds.ctx, id: ds.serviceId, typeSet: ts, metadata: types.WrapValues(defs), stateConverter: ds.stateConverter, callables: callables, states: ds.states}
}
