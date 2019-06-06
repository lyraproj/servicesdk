package service

import (
	"io"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
)

type parameter struct {
	name  string
	alias string
	typ   px.Type
	value px.Value
}

func init() {
	serviceapi.NewParameter = newParameter
}

func newParameter(name, alias string, typ px.Type, value px.Value) serviceapi.Parameter {
	if alias == name {
		alias = ``
	}
	if typ == nil {
		typ = types.DefaultRichDataType()
	}
	return &parameter{name, alias, typ, value}
}

func (p *parameter) Name() string {
	return p.name
}

func (p *parameter) Alias() string {
	return p.alias
}

func (p *parameter) Value() px.Value {
	return p.value
}

func (p *parameter) Type() px.Type {
	return p.typ
}

func (p *parameter) Get(key string) (value px.Value, ok bool) {
	switch key {
	case `name`:
		return types.WrapString(p.name), true
	case `alias`:
		if p.alias == `` {
			return px.Undef, true
		}
		return types.WrapString(p.alias), true
	case `type`:
		return p.typ, true
	case `value`:
		if p.value == nil {
			return px.Undef, true
		}
		return p.Value(), true
	case `has_value`:
		return types.WrapBoolean(p.value != nil), true
	}
	return nil, false
}

func (p *parameter) InitHash() px.OrderedMap {
	es := make([]*types.HashEntry, 0, 3)
	es = append(es, types.WrapHashEntry2(`name`, types.WrapString(p.name)))
	if p.alias != `` {
		es = append(es, types.WrapHashEntry2(`alias`, types.WrapString(p.alias)))
	}
	es = append(es, types.WrapHashEntry2(`type`, p.typ))
	if p.value != nil {
		es = append(es, types.WrapHashEntry2(`value`, p.value))
	}
	return types.WrapHash(es)
}

var ParameterMetaType px.Type

func (p *parameter) Equals(other interface{}, guard px.Guard) bool {
	return p == other
}

func (p *parameter) String() string {
	return px.ToString(p)
}

func (p *parameter) ToString(bld io.Writer, format px.FormatContext, g px.RDetect) {
	types.ObjectToString(p, format, bld, g)
}

func (p *parameter) PType() px.Type {
	return ParameterMetaType
}

func init() {
	ParameterMetaType = px.NewObjectType(`Lyra::Parameter`, `{
    attributes => {
      'name' => String,
      'type' => Type,
      'alias' => Optional[String],
      'value' => Optional[RichData],
      'has_value' => { type => Boolean, kind => derived }
    }
  }`, func(ctx px.Context, args []px.Value) px.Value {
		n := args[0].String()
		t := args[1].(px.Type)
		a := ``
		if len(args) > 2 {
			a = args[2].String()
		}
		var v px.Value
		if len(args) > 3 {
			v = args[3]
		}
		return newParameter(n, a, t, v)
	}, func(ctx px.Context, args []px.Value) px.Value {
		h := args[0].(*types.Hash)
		n := h.Get5(`name`, px.EmptyString).String()
		t := h.Get5(`type`, types.DefaultDataType()).(px.Type)
		a := ``
		if x, ok := h.Get4(`alias`); ok {
			a = x.String()
		}
		var v px.Value
		if x, ok := h.Get4(`value`); ok {
			v = x
		}
		return newParameter(n, a, t, v)
	})
}
