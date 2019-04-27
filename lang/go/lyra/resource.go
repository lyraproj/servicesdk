package lyra

import (
	"reflect"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/wf"
)

// Resource represents a declarative workflow activity
type Resource struct {
	// Name of resource. This field is mandatory
	Name string

	// When is a Condition in string form. Can be left empty
	When string

	// ExternalId can be set to the external ID of an existing resource. The resource will then
	// not be managed by Lyra.
	ExternalId string

	// Output is an optional zero value of a struct or a pointer to a struct. The exported fields
	// of that struct defines the output of the resource activity
	Output interface{}

	// State is a function that produces the desired state of the resource.
	//
	// The function can take one optional parameter which must be a struct or pointer to a struct. The exported fields of
	// that struct becomes the Input of the action.
	//
	// The function can return one or two values. The first value must be a pointer to a struct. That struct represents
	// the resource type. If an optional second value is returned, it must be of type error.
	State interface{}
}

func (r *Resource) Resolve(c px.Context, pn string) wf.Activity {
	n := r.Name
	if n == `` {
		panic(px.Error(MissingRequiredField, issue.H{`type`: `Resource`, `name`: `Name`}))
	}
	if pn != `` {
		n = pn + `::` + n
	}

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

	var input, output []px.Parameter

	// Create Output parameters from the Output struct
	if r.Output != nil {
		ov := reflect.ValueOf(r.Output)
		out := ov.Type()
		if out.Kind() == reflect.Ptr {
			out = out.Elem()
		}
		if out.Kind() != reflect.Struct {
			panic(px.Error(NotStruct, issue.H{`name`: n, `type`: out.String()}))
		}
		output = paramsFromStruct(c, out, func(name string) string {
			// Check if alias maps to a field. If it does, then the puppet name of
			// that field must be used instead
			for _, a := range ot.AttributesInfo().Attributes() {
				if a.Name() == name || a.GoName() == name {
					return a.Name()
				}
			}
			for _, f := range types.Fields(rt) {
				if f.Name == name {
					return types.FieldName(&f)
				}
			}
			panic(px.Error(px.AttributeNotFound, issue.H{`type`: rt.Name(), `name`: name}))
		})
	}

	// Create Input parameters from the state function struct parameter
	inc := ft.NumIn()
	if ft.IsVariadic() || inc > 1 {
		panic(badFunction(n, ft))
	}
	if inc == 1 {
		input = paramsFromStruct(c, ft.In(0), nil)
	}

	return wf.MakeResource(n, wf.Parse(r.When), input, output, r.ExternalId, newGoState(ot, fv, returnsError))
}
