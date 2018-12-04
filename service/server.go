package service

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/semver/semver"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wfapi"
	"reflect"
	"sync"
)

var ServerVersion = semver.MustParseVersion(`0.1.0`)

type Server struct {
	context   eval.Context
	id        eval.TypedName
	lock      sync.RWMutex
	typeSet   eval.TypeSet
	metadata  eval.List
	stateConv wfapi.StateConverter
	states    map[string]wfapi.State
	callables map[string]eval.Value
}

func (s *Server) AddApi(name string, callable interface{}) serviceapi.Definition {
	rv := reflect.ValueOf(callable)
	rt := rv.Type()
	pt, ok := s.context.ImplementationRegistry().ReflectedToType(rt)
	if !ok {
		panic(eval.Error(WF_API_TYPE_NOT_REGISTERED, issue.H{`type`: rt.Name()}))
	}

	s.lock.RLock()
	_, found := s.callables[name]
	s.lock.RUnlock()

	if found {
		panic(eval.Error(WF_ALREADY_REGISTERED, issue.H{`namespace`: eval.NsInterface, `identifier`: name}))
	}

	props := make([]*types.HashEntry, 0, 2)
	props = append(props, types.WrapHashEntry2(`interface`, pt))
	props = append(props, types.WrapHashEntry2(`style`, types.WrapString(`callable`)))
	def := serviceapi.NewDefinition(eval.NewTypedName(eval.NsDefinition, name), s.id, types.WrapHash(props))

	nmd := s.metadata.Add(def)
	defw := eval.WrapReflected(s.context, rv)

	s.lock.Lock()
	s.callables[name] = defw
	s.metadata = nmd
	s.lock.Unlock()

	return def
}


func (s *Server) State(c eval.Context, name string, input eval.OrderedMap) eval.PuppetObject {
	if s.stateConv != nil {
		s.lock.RLock()
		st, ok := s.states[name]
		s.lock.RUnlock()
		if ok {
			return s.stateConv(c, st, input)
		}
		panic(eval.Error(WF_NO_SUCH_STATE, issue.H{`name`: name}))
	}
	panic(eval.Error(WF_NO_STATE_CONVERTER, issue.H{`name`: name}))
}

func (s *Server) Identifier(eval.Context) eval.TypedName {
	return s.id
}

func (s *Server) Invoke(c eval.Context, api, name string, arguments ...eval.Value) (result eval.Value) {
	s.lock.RLock()
	iv, ok := s.callables[api]
	s.lock.RUnlock()
	if ok {
		if m, ok := iv.PType().(eval.TypeWithCallableMembers).Member(name); ok {
			defer func() {
				if x := recover(); x != nil {
					if err, ok := x.(issue.Reported); ok && string(err.Code()) == eval.EVAL_GO_FUNCTION_ERROR {
						result = eval.ErrorFromReported(c, err)
						return
					}
					panic(x)
				}
			}()
			result = m.Call(c, iv, nil, arguments)
			return
		}
	}
	panic(eval.Error(WF_NO_SUCH_METHOD, issue.H{`api`: api, `method`: name}))
}

func (s *Server) Metadata(eval.Context) (typeSet eval.TypeSet, definitions []serviceapi.Definition) {
	ds := make([]serviceapi.Definition, s.metadata.Len())
	s.lock.RLock()
	s.metadata.EachWithIndex(func(v eval.Value, i int) { ds[i] = v.(serviceapi.Definition) })
	s.lock.RUnlock()
	return s.typeSet, ds
}
