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
	tgt := px.NewTaggedType(s, nil)
	av, _ := c.Reflector().InitializerFromTagged(`Tmp`, nil, tgt).Get4(`attributes`)
	attrs := av.(px.OrderedMap)

	outCount := attrs.Len()
	params := make([]serviceapi.Parameter, 0, outCount)
	attrs.EachPair(func(k, v px.Value) {
		ad := v.(px.OrderedMap)
		tp := ad.Get5(`type`, types.DefaultAnyType()).(px.Type)
		an := k.String()
		var gn string
		if gv, ok := ad.Get4(`go_name`); ok {
			gn = gv.String()
		}
		alias := an

		var value px.Value
		if tags, ok := tgt.OtherTags()[gn]; ok {
			if v, ok := tags[`value`]; ok {
				value = types.CoerceTo(c, `value annotation`, tp, types.WrapString(v))
			} else if v, ok := tags[`lookup`]; ok {
				value = types.NewDeferred(`lookup`, types.WrapString(v))
			}
			if v, ok := tags[`alias`]; ok {
				alias = v
			}
		}

		if value == nil {
			if v, ok := ad.Get4(`value`); ok {
				// InitializerFromTagged will assign an Undef as the default value for an Optional. A Parameter
				// is however never optional unless it has an explicit value (or lookup) declared in a 'puppet'
				// tag.
				if puppetTags, ok := tgt.Tags()[gn]; ok {
					if _, ok = puppetTags.Get4(`value`); ok {
						value = v
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
