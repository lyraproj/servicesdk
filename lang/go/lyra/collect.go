package lyra

import (
	"reflect"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/wf"
)

// Collect is an step that applies another step repeatedly in parallel and
// collects the results into a slice returns variable.
type Collect struct {
	// When is a Condition in string form. Can be left empty
	When string

	// Times denotes an iteration that will happen given number of times. It is mutually exclusive
	// to Each
	//
	// The value must be either a literal integer or the zero value of a struct with one field of
	// integer type that becomes an parameters variable of the step
	Times interface{}

	// Each denotes the values to iterate over. It is mutually exclusive to Times.
	//
	// The value must be either a literal slice or the zero value of a struct with one field of
	// slice type that becomes an parameters variable of the step
	Each interface{}

	// As is the variable or variables that is the parameters of each iteration. The producer
	// must declare these variables as parameters. It must be either a single string, a slice
	// of strings, or the zero value of a struct.
	As interface{}

	// Return is the name of the slice that represents the collected data (the returns of this
	// step). The element type this slice is the returns type of the producer. Can be left empty
	// in which case the returns name is the same as the leaf name of the collect step.
	Return string

	// Step gets applied once for each iteration
	Step Step
}

func (e *Collect) Resolve(c px.Context, n string, loc issue.Location) wf.Step {
	var v px.Value
	var style wf.IterationStyle
	if e.Times != nil {
		if e.Each != nil {
			panic(px.Error(MutuallyExclusiveFields, issue.H{`fields`: []string{`Times`, `Each`}}))
		}
		v = value(c, e.Times)
		style = wf.IterationStyleTimes
	} else if e.Each == nil {
		panic(px.Error(RequireOneOfFields, issue.H{`fields`: []string{`Times`, `Each`}}))
	} else {
		v = value(c, e.Times)
		style = wf.IterationStyleEach
	}

	if e.Step == nil {
		panic(px.Error(MissingRequiredField, issue.H{`type`: `Collect`, `name`: `Producer`}))
	}

	if e.As == nil {
		panic(px.Error(MissingRequiredField, issue.H{`type`: `Collect`, `name`: `As`}))
	}

	return wf.MakeIterator(
		n, loc, wf.Parse(e.When), nil, nil, style, e.Step.Resolve(c, n, loc), v, asParams(c, e.As), issue.FirstToLower(e.Return))
}

// value is like px.Wrap but transforms single element zero element structs into parameters
func value(c px.Context, uv interface{}) px.Value {
	rv := reflect.ValueOf(uv)
	switch rv.Kind() {
	case reflect.Ptr:
		e := rv.Elem()
		if e.Len() == 1 && !e.Field(0).IsValid() {
			return paramFromStruct(c, e)
		}
	case reflect.Struct:
		if rv.Len() == 1 && !rv.Field(0).IsValid() {
			return paramFromStruct(c, rv)
		}
	case reflect.Slice:
		l := rv.Len()
		es := make([]px.Value, l)
		for i := 0; i < l; i++ {
			es[i] = value(c, rv.Index(i))
		}
		return types.WrapValues(es)
	case reflect.Map:
		ks := rv.MapKeys()
		l := len(ks)
		es := make([]*types.HashEntry, l)
		for i, k := range ks {
			es[i] = types.WrapHashEntry(value(c, k), value(c, rv.MapIndex(k)))
		}
		return types.WrapHash(es)
	}
	return px.Wrap(c, uv)
}

func paramsFromString(n string) []px.Parameter {
	return []px.Parameter{paramFromString(n)}
}

func asParams(c px.Context, ns interface{}) []px.Parameter {
	switch ns := ns.(type) {
	case string:
		return paramsFromString(ns)
	case []string:
		ps := make([]px.Parameter, len(ns))
		for i, n := range ns {
			ps[i] = paramFromString(n)
		}
		return ps
	default:
		return paramsFromStruct(c, reflect.TypeOf(ns), nil)
	}
}

func paramFromString(n string) px.Parameter {
	return px.NewParameter(issue.FirstToLower(n), types.DefaultAnyType(), nil, false)
}

func paramFromStruct(c px.Context, s reflect.Value) px.Parameter {
	params := paramsFromStruct(c, s.Type(), nil)
	if len(params) != 1 {
		panic(px.Error(NotOneStructField, issue.H{`type`: s.Type().String()}))
	}
	return params[0]
}
