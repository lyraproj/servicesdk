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
	typeSet   eval.TypeSet
	metadata  eval.List
	stateConv wfapi.StateConverter
	states    map[string]wfapi.State
	callables map[string]eval.Value
}

func (ik *Server) State(name string, input eval.OrderedMap) eval.PuppetObject {
	if ik.stateConv != nil {
		if s, ok := ik.states[name]; ok {
			return ik.stateConv(ik.context, s, input)
		}
		panic(eval.Error(WF_NO_SUCH_STATE, issue.H{`name`: name}))
	}
	panic(eval.Error(WF_NO_STATE_CONVERTER, issue.H{`name`: name}))
}

func (ik *Server) Invoke(api, name string, arguments ...eval.Value) eval.Value {
	if iv, ok := ik.callables[api]; ok {
		if m, ok := iv.PType().(eval.TypeWithCallableMembers).Member(name); ok {
			return func() (result eval.Value) {
				defer func() {
					if x := recover(); x != nil {
						if err, ok := x.(issue.Reported); ok {
							result = eval.ErrorFromReported(ik.context, err)
							return
						}
						if err, ok := x.(error); ok {
							result = eval.NewError(ik.context, err.Error(), `undefined`, eval.EVAL_FAILURE, nil, nil)
							return
						}
						panic(x)
					}
				}()
				result = m.Call(ik.context, iv, nil, arguments)
				return
			}()
		}
	}
	panic(eval.Error(WF_NO_SUCH_METHOD, issue.H{`api`: api, `method`: name}))
}

func (ik *Server) Metadata() (typeSet eval.TypeSet, definitions []serviceapi.Definition) {
	ds := make([]serviceapi.Definition, ik.metadata.Len())
	ik.metadata.EachWithIndex(func(v eval.Value, i int) { ds[i] = v.(serviceapi.Definition) })
	return ik.typeSet, ds
}
