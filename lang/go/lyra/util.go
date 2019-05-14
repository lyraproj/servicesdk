// Package lyra provides struct types that implement the Step interface. The structs can be used to declare
// a complete Lyra workflow in Golang
package lyra

import (
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/lyraproj/pcore/pcore"

	"github.com/lyraproj/servicesdk/serviceapi"

	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/grpc"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/wf"
)

// Step is implemented by Action, Resource, and Workflow
type Step interface {
	// Resolve resolves the step internals using the given Context
	Resolve(c px.Context, pn string, loc issue.Location) wf.Step
}

// Serve initializes the grpc plugin mechanism, resolves the given Step, and serves it up to the Lyra client. The
// given init function can be used to initialize a resource type package.
func Serve(n string, init func(c px.Context), a Step) {
	// Configuring hclog like this allows Lyra to handle log levels automatically
	hclog.DefaultOptions = &hclog.LoggerOptions{
		Name:            "Go",
		Level:           hclog.LevelFromString(os.Getenv("LYRA_LOG_LEVEL")),
		JSONFormat:      true,
		IncludeLocation: false,
		Output:          os.Stderr,
	}
	// Tell issue reporting to amend all errors with a stack trace.
	issue.IncludeStacktrace(hclog.DefaultOptions.Level <= hclog.Debug)

	_, file, _, _ := runtime.Caller(1) // Assume Step declaration resides in Caller
	loc := issue.NewLocation(file, 0, 0)

	pcore.Do(func(c px.Context) {
		c.DoWithLoader(service.FederatedLoader(c.Loader()), func() {
			if init != nil {
				init(c)
			}
			sb := service.NewServiceBuilder(c, `Step::Service::`+strings.Title(n))
			sb.RegisterStateConverter(StateConverter)
			sb.RegisterStep(a.Resolve(c, n, loc))
			grpc.Serve(c, sb.Server())
		})
	})
}

// StringPtr returns a pointer to the given string. Useful when a pointer to a literal string is needed
func StringPtr(s string) *string {
	return &s
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func reflectParameters(ctx px.Context, param reflect.Type, parameters px.OrderedMap) reflect.Value {
	ptr := param.Kind() == reflect.Ptr
	if ptr {
		param = param.Elem()
	}
	in := reflect.New(param).Elem()
	t := in.NumField()
	r := ctx.Reflector()
	for i := 0; i < t; i++ {
		pn := issue.FirstToLower(param.Field(i).Name)
		r.ReflectTo(parameters.Get5(pn, px.Undef), in.Field(i))
	}
	if ptr {
		in = in.Addr()
	}
	return in
}

func badFunction(name string, typ reflect.Type) error {
	return px.Error(BadFunction, issue.H{`name`: name, `type`: typ.String()})
}

func ParametersFromGoStruct(c px.Context, v interface{}) []serviceapi.Parameter {
	if v == nil {
		return nil
	}
	return paramsFromStruct(c, reflect.TypeOf(v), nil)
}

func paramsFromStruct(c px.Context, s reflect.Type, nameMapper func(string) string) []serviceapi.Parameter {
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	if s.Kind() != reflect.Struct {
		panic(px.Error(NotStruct, issue.H{`type`: s.String()}))
	}
	av, _ := c.Reflector().InitializerFromTagged(`Tmp`, nil, px.NewTaggedType(s, nil)).Get4(`attributes`)
	attrs := av.(px.OrderedMap)

	outCount := attrs.Len()
	params := make([]serviceapi.Parameter, 0, outCount)
	var value px.Value
	attrs.EachPair(func(k, v px.Value) {
		ad := v.(px.OrderedMap)
		tp := ad.Get5(`type`, types.DefaultAnyType()).(px.Type)
		an := k.String()
		alias := an
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
					if v, ok := tm.Get4(`alias`); ok {
						alias = v.String()
					}
				}
			}
		}

		if nameMapper != nil {
			alias = nameMapper(alias)
		}
		params = append(params, serviceapi.NewParameter(an, alias, tp, value))
	})
	return params
}
