package lyra

import (
	"io"
	"reflect"

	"github.com/lyraproj/servicesdk/serviceapi"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/wf"
)

// Action is an imperative workflow step
type Action struct {
	// When is a Condition in string form. Can be left empty
	When string

	// Do is the actual function that is executed by this action.
	//
	// The function can take one optional parameter which must be a struct or pointer to a struct. The exported fields of
	// that struct becomes the Parameters of the action.
	//
	// The function can return zero, one, or two values. If one value is returned, that value can be either an error, a
	// struct, or a pointer to a struct. If two values are returned, the first value must be struct or a pointer to a
	// struct and the second must be an error. The exported fields of a returned struct becomes the returns of the action.
	Do interface{}
}

func (a *Action) Resolve(c px.Context, n string, loc issue.Location) wf.Step {
	fv := reflect.ValueOf(a.Do)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		panic(px.Error(NotActionFunction, issue.H{`name`: n, `type`: ft.String()}))
	}

	var parameters, returns []serviceapi.Parameter
	inc := ft.NumIn()
	if inc > 0 {
		idx := 0
		if ft.In(0).AssignableTo(px.ContextType) {
			inc--
			idx++
		}
		if inc == 1 {
			parameters = paramsFromStruct(c, ft.In(idx), nil)
		}
	}

	if ft.IsVariadic() || inc > 1 {
		panic(badFunction(n, ft))
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
			returns = paramsFromStruct(c, ft.Out(0), nil)
		}
	case 2:
		// First return type must be a struct, second must be an error
		returnsError = ft.Out(1).AssignableTo(errorInterface)
		if !returnsError {
			panic(badFunction(n, ft))
		}
		returns = paramsFromStruct(c, ft.Out(0), nil)
	default:
		panic(badFunction(n, ft))
	}

	ga := &goAction{returnsError: returnsError, doer: fv}
	as := wf.MakeAction(n, loc, wf.Parse(a.When), parameters, returns, ga)
	ga.action = as
	return as
}

type goAction struct {
	action       wf.Action
	doer         reflect.Value
	returnsError bool
}

var goActionType px.ObjectType

func init() {
	goActionType = px.NewGoObjectType(`Lyra::Action`, reflect.TypeOf(&goAction{}), `{
		functions => {
      do => Callable[[Hash[String,RichData]], Hash[String,RichData]]
    }
  }`)
}

// Call checks if the method is 'do' and then converts the single argument OrderedMap into the go struct required by the
// go function, calls the function, and then converts the returned go struct into an OrderedMap which is returned.
// Call will return nil, false for any other method than 'do'
func (a *goAction) Call(ctx px.Context, method px.ObjFunc, args []px.Value, block px.Lambda) (px.Value, bool) {
	if method.Name() != `do` {
		return nil, false
	}
	fvType := a.doer.Type()

	parameters := args[0].(px.OrderedMap)
	params := make([]reflect.Value, 0)
	if fvType.NumIn() > 0 {
		inType := fvType.In(0)
		if inType.AssignableTo(px.ContextType) {
			params = append(params, reflect.ValueOf(ctx))
			if fvType.NumIn() > 1 {
				params = append(params, reflectParameters(ctx, fvType.In(1), parameters))
			}
		} else {
			params = append(params, reflectParameters(ctx, inType, parameters))
		}
	}

	defer a.amendError()

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
		panic(px.Error(NotStruct, issue.H{`type`: rt.String()}))
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
	return goActionType
}

func (a *goAction) Get(key string) (value px.Value, ok bool) {
	return nil, false
}

func (a *goAction) InitHash() px.OrderedMap {
	return px.EmptyMap
}

func (a *goAction) amendError() {
	if r := recover(); r != nil {
		if rx, ok := r.(issue.Reported); ok {
			// Location and stack included in nested error
			r = issue.ErrorWithStack(wf.ActionExecutionError, issue.H{`step`: a.action.Label()}, nil, rx, ``)
		} else {
			r = issue.NewNested(wf.ActionExecutionError, issue.H{`step`: a.action.Label()}, a.action.Origin(), wf.ToError(r))
		}
		panic(r)
	}
}
