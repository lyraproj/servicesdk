package service

import (
	"errors"
	"io"
	"os"
	"reflect"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
)

var ErrorMetaType px.ObjectType

func init() {
	ErrorMetaType = px.NewGoObjectType(`Error`, reflect.TypeOf((*serviceapi.ErrorObject)(nil)).Elem(), `{
		type_parameters => {
		  kind => Optional[Variant[String,Regexp,Type[Enum],Type[Pattern],Type[NotUndef],Type[Undef]]],
	  	issue_code => Optional[Variant[String,Regexp,Type[Enum],Type[Pattern],Type[NotUndef],Type[Undef]]]
		},
		attributes => {
		  message => String[1],
	  	kind => { type => Optional[String[1]], value => undef },
		  issue_code => { type => Optional[String[1]], value => undef },
		  partial_result => { type => Data, value => undef },
	  	details => { type => Optional[Hash[String[1],RichData]], value => {} },
		}}`,
		func(ctx px.Context, args []px.Value) px.Value {
			return newError2(ctx, args...)
		},
		func(ctx px.Context, args []px.Value) px.Value {
			return newErrorFromHash(ctx, args[0].(px.OrderedMap))
		})

	serviceapi.NewError = newError
	serviceapi.ErrorFromReported = errorFromReported
}

type errorObj struct {
	typ           px.Type
	message       string
	kind          string
	issueCode     string
	partialResult px.Value
	details       px.OrderedMap
}

func newError2(c px.Context, args ...px.Value) serviceapi.ErrorObject {
	argc := len(args)
	ev := &errorObj{partialResult: px.Undef, details: px.EmptyMap}
	ev.message = args[0].String()
	if argc > 1 {
		ev.kind = args[1].String()
		if argc > 2 {
			ev.issueCode = args[2].String()
			if argc > 3 {
				ev.partialResult = args[3]
				if argc > 4 {
					ev.details = args[4].(px.OrderedMap)
				}
			}
		}
	}
	ev.initType(c)
	return ev
}

func newError(c px.Context, message, kind, issueCode string, partialResult px.Value, details px.OrderedMap) serviceapi.ErrorObject {
	if partialResult == nil {
		partialResult = px.Undef
	}
	if details == nil {
		details = px.EmptyMap
	}
	ev := &errorObj{message: message, kind: kind, issueCode: issueCode, partialResult: partialResult, details: details}
	ev.initType(c)
	return ev
}

func errorFromReported(c px.Context, err issue.Reported) serviceapi.ErrorObject {
	ev := &errorObj{partialResult: px.Undef, details: px.EmptyMap}
	ev.message = err.Error()
	ev.kind = `PUPPET_ERROR`
	ev.issueCode = string(err.Code())
	ds := make([]*types.HashEntry, 0)
	if loc := err.Location(); loc != nil {
		ds = append(ds, types.WrapHashEntry2(`location`, types.WrapString(issue.LocationString(loc))))
	}
	keys := err.Keys()
	if len(keys) > 0 {
		args := make([]*types.HashEntry, len(keys))
		for i, k := range keys {
			av := err.Argument(k)
			var arg px.Value
			if ea, ok := av.(error); ok {
				arg = types.WrapString(ea.Error())
			} else {
				arg = px.Wrap(c, av)
			}
			args[i] = types.WrapHashEntry2(k, arg)
		}
		ds = append(ds, types.WrapHashEntry2(`arguments`, types.WrapHash(args)))
	}
	if cause := err.Cause(); cause != nil {
		var cv px.Value
		if cr, ok := cause.(issue.Reported); ok {
			cv = errorFromReported(c, cr)
		} else {
			cv = types.WrapString(cause.Error())
		}
		ds = append(ds, types.WrapHashEntry2(`cause`, cv))
	}
	stack := err.Stack()
	if stack != `` {
		ds = append(ds, types.WrapHashEntry2(`stack`, types.WrapString(stack)))
	}
	if host, err := os.Hostname(); err == nil {
		ds = append(ds, types.WrapHashEntry2(`host`, types.WrapString(host)))
	}
	if exec, err := os.Executable(); err == nil {
		ds = append(ds, types.WrapHashEntry2(`executable`, types.WrapString(exec)))
	}
	if len(ds) > 0 {
		ev.details = types.WrapHash(ds)
	}
	ev.initType(c)
	return ev
}

func newErrorFromHash(c px.Context, hash px.OrderedMap) serviceapi.ErrorObject {
	ev := &errorObj{}
	ev.message = hash.Get5(`message`, px.EmptyString).String()
	ev.kind = hash.Get5(`kind`, px.EmptyString).String()
	ev.issueCode = hash.Get5(`issue_code`, px.EmptyString).String()
	ev.partialResult = hash.Get5(`partial_result`, px.Undef)
	ev.details = hash.Get5(`details`, px.EmptyMap).(px.OrderedMap)
	ev.initType(c)
	return ev
}

func (e *errorObj) Details() px.OrderedMap {
	return e.details
}

func (e *errorObj) IssueCode() string {
	return e.issueCode
}

func (e *errorObj) Kind() string {
	return e.kind
}

func (e *errorObj) Message() string {
	return e.message
}

func (e *errorObj) PartialResult() px.Value {
	return e.partialResult
}

func (e *errorObj) ToReported() (issue.Reported, bool) {
	code := issue.Code(e.issueCode)
	if _, ok := issue.ForCode2(code); ok {
		args := issue.NoArgs
		var loc issue.Location
		var cause error
		stack := ``
		if e.details.Len() > 0 {
			if ls, ok := e.details.Get4(`location`); ok {
				loc = issue.ParseLocation(ls.String())
			}
			if am, ok := e.details.Get4(`arguments`); ok {
				if ah, ok := am.(px.OrderedMap); ok {
					args = make(issue.H, ah.Len())
					ah.EachPair(func(k, v px.Value) {
						args[k.String()] = v
					})
				}
			}
			if cs, ok := e.details.Get4(`cause`); ok {
				if cse, ok := cs.(serviceapi.ErrorObject); ok {
					if cr, ok := cse.ToReported(); ok {
						cause = cr
					} else {
						cause = errors.New(cse.Message())
					}
				} else {
					cause = errors.New(cs.String())
				}
			}

			if ss, ok := e.details.Get4(`stack`); ok {
				stack = ss.String()
			}
		}
		return issue.ErrorWithStack(code, args, loc, cause, stack), true
	}

	// Code does not represent a valid issue.
	return nil, false
}

func (e *errorObj) String() string {
	return px.ToString(e)
}

func (e *errorObj) Equals(other interface{}, guard px.Guard) bool {
	if o, ok := other.(*errorObj); ok {
		return e.message == o.message && e.kind == o.kind && e.issueCode == o.issueCode &&
			px.Equals(e.partialResult, o.partialResult, guard) &&
			px.Equals(e.details, o.details, guard)
	}
	return false
}

func (e *errorObj) ToString(b io.Writer, s px.FormatContext, g px.RDetect) {
	types.ObjectToString(e, s, b, g)
}

func (e *errorObj) PType() px.Type {
	return e.typ
}

func (e *errorObj) Get(key string) (value px.Value, ok bool) {
	switch key {
	case `message`:
		return types.WrapString(e.message), true
	case `kind`:
		if e.kind == `` {
			return px.Undef, true
		}
		return types.WrapString(e.kind), true
	case `issue_code`:
		if e.issueCode == `` {
			return px.Undef, true
		}
		return types.WrapString(e.issueCode), true
	case `partial_result`:
		return e.partialResult, true
	case `details`:
		return e.details, true
	default:
		return nil, false
	}
}

func (e *errorObj) InitHash() px.OrderedMap {
	return ErrorMetaType.InstanceHash(e)
}

func (e *errorObj) initType(c px.Context) {
	if e.details == nil {
		e.details = px.EmptyMap
	}
	if e.kind == `` && e.issueCode == `` {
		e.typ = ErrorMetaType
	} else {
		params := make([]*types.HashEntry, 0)
		if e.kind != `` {
			params = append(params, types.WrapHashEntry2(`kind`, types.WrapString(e.kind)))
		}
		if e.issueCode != `` {
			params = append(params, types.WrapHashEntry2(`issue_code`, types.WrapString(e.issueCode)))
		}
		e.typ = types.NewObjectTypeExtension(c, ErrorMetaType, []px.Value{types.WrapHash(params)})
	}
}
