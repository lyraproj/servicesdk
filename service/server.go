package service

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-semver/semver"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
)

var ServerVersion = semver.MustParseVersion(`0.1.0`)

type Server struct {
	context   eval.Context
	typeSet   eval.TypeSet
	metadata  eval.List
	callables map[string]eval.Value
}

func (ik *Server) Invoke(api, name string, arguments ...eval.Value) eval.Value {
	if iv, ok := ik.callables[api]; ok {
		if m, ok := iv.PType().(eval.TypeWithCallableMembers).Member(name); ok {
			return m.Call(ik.context, iv, nil, arguments)
		}
	}
	panic(eval.Error(WF_NO_SUCH_METHOD, issue.H{`api`: api, `method`: name}))
}

func (m *Server) Metadata() (typeSet eval.TypeSet, definitions []serviceapi.Definition) {
	ds := make([]serviceapi.Definition, m.metadata.Len())
	m.metadata.EachWithIndex(func(v eval.Value, i int) { ds[i] = v.(serviceapi.Definition) })
	return m.typeSet, ds
}
