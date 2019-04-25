package wf

import (
	"io"
	"reflect"

	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
)

const (
	WfNotFunction = `WF_NOT_FUNCTION`
	WfBadFunction = `WF_BAD_FUNCTION`
	WfNotStruct   = `WF_NOT_STRUCT`
)

type goAction struct {
	Action
	doer         reflect.Value
	returnsError bool
	resolved     bool
}

var GoActionType px.ObjectType

func init() {
	GoActionType = px.NewGoObjectType(`Lyra::Action`, reflect.TypeOf(&goAction{}), `{
		functions => {
      do => Callable[[Hash[String,RichData]], Hash[String,RichData]]
    }
  }`)
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// Call checks if the method is 'do' and then converts the single argument OrderedMap into the go struct required by the
// go function, calls the function, and then converts the returned go struct into an OrderedMap which is returned.
// Call will return nil, false for any other method than 'do'
func (a *goAction) Call(ctx px.Context, method px.ObjFunc, args []px.Value, block px.Lambda) (px.Value, bool) {
	if method.Name() != `do` {
		return nil, false
	}
	fvType := a.doer.Type()

	input := args[0].(px.OrderedMap)
	params := make([]reflect.Value, 0)
	if fvType.NumIn() > 0 {
		inType := fvType.In(0)
		ptr := inType.Kind() == reflect.Ptr
		if ptr {
			inType = inType.Elem()
		}
		in := reflect.New(inType).Elem()
		t := in.NumField()
		r := ctx.Reflector()
		for i := 0; i < t; i++ {
			pn := issue.FirstToLower(inType.Field(i).Name)
			r.ReflectTo(input.Get5(pn, px.Undef), in.Field(i))
		}
		if ptr {
			in = in.Addr()
		}
		params = append(params, in)
	}

	result := a.doer.Call(params)
	var re, rs reflect.Value
	switch len(result) {
	case 1:
		rs = result[0]
		if a.returnsError {
			re = result[0]
		}
	case 2:
		rs = result[0]
		if a.returnsError {
			re = result[1]
		}
	}

	if re.IsValid() && re.Type().AssignableTo(errorInterface) {
		panic(rs.Interface())
	}

	if !rs.IsValid() {
		return px.EmptyMap, true
	}

	rt := rs.Type()
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rs = rs.Elem()
	}
	if rt.Kind() != reflect.Struct {
		panic(px.Error(WfNotStruct, issue.H{`type`: rt.String()}))
	}
	fc := rt.NumField()
	entries := make([]*types.HashEntry, fc)
	for i := 0; i < fc; i++ {
		ft := rt.Field(i)
		v := rs.Field(i)
		n := issue.FirstToLower(ft.Name)
		if v.IsValid() {
			entries[i] = types.WrapHashEntry2(n, px.Wrap(ctx, v))
		} else {
			entries[i] = types.WrapHashEntry2(n, px.Undef)
		}
	}
	return types.WrapHash(entries), true
}

func (a *goAction) String() string {
	return px.ToString(a)
}

func (a *goAction) Equals(value interface{}, guard px.Guard) bool {
	return a == value
}

func (a *goAction) ToString(bld io.Writer, format px.FormatContext, g px.RDetect) {
	types.ObjectToString(a, format, bld, g)
}

func (a *goAction) PType() px.Type {
	return GoActionType
}

func (a *goAction) Get(key string) (value px.Value, ok bool) {
	return nil, false
}

func (a *goAction) InitHash() px.OrderedMap {
	return px.EmptyMap
}

// GoAction creates an Action from a Go function with the given name. The function takes zero arguments or one argument in the
// form of a struct. The function must either have no return or return either a pointer to a struct, an error, or both.
//
// The "input" declaration for the activity is reflected from the the fields in the struct argument and the
// "output" declaration is reflected from the fields in the returned struct.
func GoAction(name, when string, function interface{}) Action {
	fv := reflect.ValueOf(function)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		panic(px.Error(WfNotFunction, issue.H{`name`: name, `type`: ft.String()}))
	}

	var cond Condition = Always
	if when != `` {
		cond = Parse(when)
	}

	ga := &goAction{doer: fv, resolved: false}
	a := MakeAction(name, cond, nil, nil, ga)
	ga.Action = a
	return ga
}

func (a *goAction) Resolve(c px.Context) {
	if a.resolved {
		return
	}

	hclog.Default().Debug("GoAction.Resolve()", "name", a.Name())

	var input, output []px.Parameter
	ft := a.doer.Type()
	inc := ft.NumIn()
	if ft.IsVariadic() || inc > 1 {
		panic(badActionFunction(a.Label(), ft))
	}

	oc := ft.NumOut()
	returnsError := false
	switch oc {
	case 0:
		// OK. Function can evaluate and not produce anything
	case 1:
		// Return type must be an error or a struct
		returnsError = ft.Out(0).AssignableTo(errorInterface)
		if !returnsError {
			output = paramsFromStruct(c, ft.Out(0))
		}
	case 2:
		// First return type must be a struct, second must be an error
		returnsError = ft.Out(1).AssignableTo(errorInterface)
		if !returnsError {
			panic(badActionFunction(a.Label(), ft))
		}
		output = paramsFromStruct(c, ft.Out(0))
	default:
		panic(badActionFunction(a.Label(), ft))
	}

	if inc == 1 {
		input = paramsFromStruct(c, ft.In(0))
	}
	a.returnsError = returnsError
	a.Action = MakeAction(a.Name(), a.When(), input, output, a)
	a.resolved = true
}

func badActionFunction(name string, typ reflect.Type) error {
	return px.Error(WfBadFunction, issue.H{`name`: name, `type`: typ.String()})
}

func ParametersFromGoStruct(c px.Context, v interface{}) []px.Parameter {
	if v == nil {
		return nil
	}
	return paramsFromStruct(c, reflect.TypeOf(v))
}

func paramsFromStruct(c px.Context, s reflect.Type) []px.Parameter {
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	if s.Kind() != reflect.Struct {
		panic(px.Error(WfNotStruct, issue.H{`type`: s.String()}))
	}
	av, _ := c.Reflector().InitializerFromTagged(`Tmp`, nil, px.NewTaggedType(s, nil)).Get4(`attributes`)
	attrs := av.(px.OrderedMap)

	outCount := attrs.Len()
	params := make([]px.Parameter, 0, outCount)
	var value px.Value
	attrs.EachPair(func(k, v px.Value) {
		ad := v.(px.OrderedMap)
		tp := ad.Get5(`type`, types.DefaultAnyType()).(px.Type)
		if v, ok := ad.Get4(`value`); ok {
			value = v
		} else {
			if an, ok := ad.Get4(`annotations`); ok {
				if tags, ok := an.(px.OrderedMap).Get(types.TagsAnnotationType); ok {
					tm := tags.(px.OrderedMap)
					if v, ok := tm.Get4(`value`); ok {
						value = types.CoerceTo(c, `value annotation`, tp, v)
					} else if v, ok := tm.Get4(`lookup`); ok {
						value = types.NewDeferred(`lookup`, v)
					}
				}
			}
		}
		params = append(params, px.NewParameter(k.String(), tp, value, false))
	})
	return params
}
