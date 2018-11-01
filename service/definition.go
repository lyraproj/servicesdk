package service

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"io"
)

// Identifier TypedName namespaces. Used by a service to identify what the type of entity to look for.

// Interface denotes an entity that must have an "interface" property that appoints
// an object type which in turn contains a declaration of the methods that the interface
// implements.
const Interface = eval.Namespace(`interface`)

// Activity denotes an entity that can participate in a workflow. The entity must
// declare input and output parameters. An activity of type "action" may also be an interface
// in which case it must have an "interface" property
const Activity = eval.Namespace(`activity`)

// ServiceId TypedName namespaces. Used by the Loader to determine the right type
// of RPC mechanism to use when communicating with the service.

// Plugin denotes a service that is a Hashicorp go-plugin
const Plugin = eval.Namespace(`plugin`)

// RESTFul denotes a service that is a RESTFul http or https service.
const RESTFul = eval.Namespace(`RESTFul`)

type Definition interface {
	eval.Value

	// Identifier returns a TypedName that uniquely identifies the activity within the service.
	Identifier() eval.TypedName

	// ServiceId is the identifier of the service
	ServiceId() eval.TypedName

	// Properties is an ordered map of properties of this definition. Will be of type
	// Hash[Pattern[/\A[a-z][A-Za-z]+\z/],RichData]
	Properties() eval.OrderedMap
}

var Definition_Type eval.Type

func init() {
	Definition_Type = eval.NewObjectType(`Service::Definition`, `{
    attributes => {
      identity => TypedName,
      serviceId => TypedName,
      properties => Hash[String,RichData]
    }
  }`,

	func(ctx eval.Context, args []eval.Value) eval.Value {
		identity := args[0].(eval.TypedName)
		service_id := args[1].(eval.TypedName)
		properties := args[2].(eval.OrderedMap)
		return NewDefinition(identity, service_id, properties)
	},

	func(ctx eval.Context, args []eval.Value) eval.Value {
		h := args[0].(*types.HashValue)
		identity := h.Get5(`identity`, eval.UNDEF).(eval.TypedName)
		service_id := h.Get5(`serviceId`, eval.UNDEF).(eval.TypedName)
		properties := h.Get5(`properties`, eval.EMPTY_MAP).(eval.OrderedMap)
		return NewDefinition(identity, service_id, properties)
	})
}

func NewDefinition(identity, serviceId eval.TypedName, properties eval.OrderedMap) Definition {
	return &definition{identity, serviceId, properties}
}

type definition struct {
	identifier eval.TypedName
	serviceId  eval.TypedName
	properties eval.OrderedMap
}

func (d *definition) Get(key string) (value eval.Value, ok bool) {
	switch key {
	case `identifier`:
		return d.identifier, true
	case `serviceId`:
		return d.serviceId, true
	case `properties`:
		return d.properties, true
	}
	return nil, false
}

func (d *definition) InitHash() eval.OrderedMap {
	es := make([]*types.HashEntry, 0, 3)
	es = append(es, types.WrapHashEntry2(`identifier`, d.identifier))
	es = append(es, types.WrapHashEntry2(`serviceId`, d.serviceId))
	es = append(es, types.WrapHashEntry2(`properties`, d.properties))
	return types.WrapHash(es)
}

func (d *definition) Equals(other interface{}, g eval.Guard) bool {
	if o, ok := other.(*definition); ok {
		return d.identifier == o.identifier && d.serviceId.Equals(o.serviceId, g) && d.properties.Equals(o.properties, g)
	}
	return false
}

func (d *definition) Identifier() eval.TypedName {
	return d.identifier
}

func (d *definition) ServiceId() eval.TypedName {
	return d.serviceId
}

func (d *definition) Properties() eval.OrderedMap {
	return d.properties
}

func (d *definition) String() string {
	return eval.ToString(d)
}

func (d *definition) ToString(bld io.Writer, format eval.FormatContext, g eval.RDetect) {
	types.ObjectToString(d, format, bld, g)
}

func (d *definition) PType() eval.Type {
	return Definition_Type
}
