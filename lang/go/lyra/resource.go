package lyra

import (
	"reflect"
	"strings"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/wf"
)

// Resource represents a declarative workflow step
type Resource struct {
	// When is a Condition in string form. Can be left empty
	When string

	// ExternalId can be set to the external ID of an existing resource. The resource will then
	// not be managed by Lyra.
	ExternalId string

	// Return is an optional zero value of a struct or a pointer to a struct. The exported fields
	// of that struct defines the returns of the resource step
	Return interface{}

	// State is a function that produces the desired state of the resource.
	//
	// The function can take one optional parameter which must be a struct or pointer to a struct. The exported fields of
	// that struct becomes the Parameters of the action.
	//
	// The function can return one or two values. The first value must be a pointer to a struct. That struct represents
	// the resource type. If an optional second value is returned, it must be of type error.
	State interface{}
}

func (r *Resource) Resolve(c px.Context, n string) wf.Step {
	fv := reflect.ValueOf(r.State)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		panic(px.Error(NotStateFunction, issue.H{`name`: n, `type`: ft.String()}))
	}

	// Derive the resource type from the state function return value
	var rt reflect.Type
	oc := ft.NumOut()
	returnsError := false
	switch oc {
	case 0:
		// Function must return a resource evaluate and not produce anything
		panic(badFunction(n, ft))
	case 1:
		// Return type must a struct
		rt = ft.Out(0)
	case 2:
		// First return type must be a struct, second must be an error
		returnsError = ft.Out(1).AssignableTo(errorInterface)
		if !returnsError {
			panic(badFunction(n, ft))
		}
		rt = ft.Out(0)
	}

	var t px.Type
	var ot px.ObjectType

	ok := false
	if rt != nil {
		switch rt.Kind() {
		case reflect.Ptr:
			if rt.Elem().Kind() == reflect.Struct {
				t, ok = c.ImplementationRegistry().ReflectedToType(rt)
			}
		case reflect.Struct:
			rt = reflect.PtrTo(rt)
			t, ok = c.ImplementationRegistry().ReflectedToType(rt)
		}
		if ok {
			ot, ok = t.(px.ObjectType)
		}
	}
	if !ok {
		panic(badFunction(n, ft))
	}

	var parameters, returns []px.Parameter

	// Create return parameters from the Returns struct
	if r.Return != nil {
		ov := reflect.ValueOf(r.Return)
		out := ov.Type()
		if out.Kind() == reflect.Ptr {
			out = out.Elem()
		}
		if out.Kind() != reflect.Struct {
			panic(px.Error(NotStruct, issue.H{`name`: n, `type`: out.String()}))
		}
		returns = paramsFromStruct(c, out, func(name string) string {
			// Check if alias maps to a field. If it does, then the puppet name of
			// that field must be used instead
			for _, a := range ot.AttributesInfo().Attributes() {
				if strings.EqualFold(a.Name(), name) || strings.EqualFold(a.GoName(), name) {
					return a.Name()
				}
			}
			for _, f := range types.Fields(rt) {
				if strings.EqualFold(f.Name, name) {
					return types.FieldName(&f)
				}
			}
			panic(px.Error(px.AttributeNotFound, issue.H{`type`: ot.Name(), `name`: name}))
		})
	}

	// Create Parameters parameters from the state function struct parameter
	inc := ft.NumIn()
	if ft.IsVariadic() || inc > 1 {
		panic(badFunction(n, ft))
	}
	if inc == 1 {
		parameters = paramsFromStruct(c, ft.In(0), nil)
	}

	return wf.MakeResource(n, wf.Parse(r.When), parameters, returns, r.ExternalId, newGoState(ot, fv, returnsError))
}
