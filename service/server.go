package service

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-semver/semver"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/puppetlabs/go-servicesdk/wfapi"
)

var ServerVersion = semver.MustParseVersion(`0.1.0`)

type Server struct {
	context   eval.Context
	id        eval.TypedName
	typeSet   eval.TypeSet
	metadata  eval.List
	stateConv wfapi.StateConverter
	states    map[string]wfapi.State
	callables map[string]eval.Value
}

func (s *Server) State(name string, input eval.OrderedMap) eval.PuppetObject {
	if s.stateConv != nil {
		if st, ok := s.states[name]; ok {
			return s.stateConv(s.context.Fork(), st, input)
		}
		panic(eval.Error(WF_NO_SUCH_STATE, issue.H{`name`: name}))
	}
	panic(eval.Error(WF_NO_STATE_CONVERTER, issue.H{`name`: name}))
}

func (s *Server) Identifier() eval.TypedName {
	return s.id
}

func (s *Server) Invoke(api, name string, arguments ...eval.Value) eval.Value {
	if iv, ok := s.callables[api]; ok {
		if m, ok := iv.PType().(eval.TypeWithCallableMembers).Member(name); ok {
			return func() (result eval.Value) {
				c := s.context.Fork()
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
			}()
		}
	}
	panic(eval.Error(WF_NO_SUCH_METHOD, issue.H{`api`: api, `method`: name}))
}

func (s *Server) Metadata() (typeSet eval.TypeSet, definitions []serviceapi.Definition) {
	ds := make([]serviceapi.Definition, s.metadata.Len())
	s.metadata.EachWithIndex(func(v eval.Value, i int) { ds[i] = v.(serviceapi.Definition) })
	return s.typeSet, ds
}
