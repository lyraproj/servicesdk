package service

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"io"
)

var Definition_Type eval.Type

func init() {
	Definition_Type = eval.NewObjectType(`Service::Definition`, `{
    attributes => {
      identifier => TypedName,
      serviceId => TypedName,
      properties => Hash[String,RichData]
    }
  }`,

		func(ctx eval.Context, args []eval.Value) eval.Value {
			identifier := args[0].(eval.TypedName)
			service_id := args[1].(eval.TypedName)
			properties := args[2].(eval.OrderedMap)
			return newDefinition(identifier, service_id, properties)
		},

		func(ctx eval.Context, args []eval.Value) eval.Value {
			h := args[0].(*types.HashValue)
			identifier := h.Get5(`identifier`, eval.UNDEF).(eval.TypedName)
			service_id := h.Get5(`serviceId`, eval.UNDEF).(eval.TypedName)
			properties := h.Get5(`properties`, eval.EMPTY_MAP).(eval.OrderedMap)
			return newDefinition(identifier, service_id, properties)
		})

	serviceapi.NewDefinition = newDefinition
}

func newDefinition(identifier, serviceId eval.TypedName, properties eval.OrderedMap) serviceapi.Definition {
	return &definition{identifier, serviceId, properties}
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
