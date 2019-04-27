package lyra

import (
	"reflect"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wf"
)

type goState struct {
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
	return &goState{resourceType, stateFunc, returnsError}
}

func (s *goState) call(c px.Context, input px.OrderedMap) px.PuppetObject {
	fv := s.stateFunc
	fvType := fv.Type()
	var params []reflect.Value
	if fvType.NumIn() == 1 {
		params = []reflect.Value{reflectInput(c, fvType.In(0), input)}
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

func StateConverter(c px.Context, state wf.State, input px.OrderedMap) px.PuppetObject {
	return state.(*goState).call(c, input)
}
