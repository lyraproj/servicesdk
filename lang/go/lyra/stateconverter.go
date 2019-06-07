package lyra

import (
	"reflect"

	"github.com/lyraproj/issue/issue"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wf"
)

type goState struct {
	resource     wf.Resource
	resourceType px.ObjectType
	stateFunc    reflect.Value
	returnsError bool
}

func (s *goState) Type() px.ObjectType {
	return s.resourceType
}

func (s *goState) State() interface{} {
	return s.stateFunc
}

func newGoState(resourceType px.ObjectType, stateFunc reflect.Value, returnsError bool) *goState {
	return &goState{nil, resourceType, stateFunc, returnsError}
}

func (s *goState) call(c px.Context, parameters px.OrderedMap) px.PuppetObject {
	defer s.amendError()

	fv := s.stateFunc
	fvType := fv.Type()
	var params []reflect.Value
	if fvType.NumIn() == 1 {
		params = []reflect.Value{reflectParameters(c, fvType.In(0), parameters)}
	}
	result := fv.Call(params)
	var re, rs reflect.Value
	switch len(result) {
	case 1:
		rs = result[0]
		if s.returnsError {
			re = result[0]
		}
	case 2:
		rs = result[0]
		if s.returnsError {
			re = result[1]
		}
	}
	if re.IsValid() && re.Type().AssignableTo(errorInterface) {
		panic(rs.Interface())
	}
	return px.WrapReflected(c, rs).(px.PuppetObject)
}

func StateConverter(c px.Context, state wf.State, parameters px.OrderedMap) px.PuppetObject {
	return state.(*goState).call(c, parameters)
}

func (a *goState) amendError() {
	if r := recover(); r != nil {
		if rx, ok := r.(issue.Reported); ok {
			// Location and stack included in nested error
			r = issue.ErrorWithStack(wf.StateCreationError, issue.H{`step`: a.resource.Label()}, nil, rx, ``)
		} else {
			r = issue.NewNested(wf.StateCreationError, issue.H{`step`: a.resource.Label()}, a.resource.Origin(), wf.ToError(r))
		}
		panic(r)
	}
}
