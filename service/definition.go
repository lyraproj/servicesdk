package service

import (
	"fmt"
	"io"
	"reflect"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
)

func init() {
	serviceapi.DefinitionMetaType = px.NewGoObjectType(`Service::Definition`, reflect.TypeOf((*serviceapi.Definition)(nil)).Elem(), `{
    attributes => {
      identifier => TypedName,
      serviceId => TypedName,
      properties => Hash[String,RichData]
    }
  }`,

		func(ctx px.Context, args []px.Value) px.Value {
			identifier := args[0].(px.TypedName)
			serviceId := args[1].(px.TypedName)
			properties := args[2].(px.OrderedMap)
			return newDefinition(identifier, serviceId, properties)
		},

		func(ctx px.Context, args []px.Value) px.Value {
			h := args[0].(*types.Hash)
			identifier := h.Get5(`identifier`, px.Undef).(px.TypedName)
			serviceId := h.Get5(`serviceId`, px.Undef).(px.TypedName)
			properties := h.Get5(`properties`, px.EmptyMap).(px.OrderedMap)
			return newDefinition(identifier, serviceId, properties)
		})

	serviceapi.NewDefinition = newDefinition
}

func newDefinition(identifier, serviceId px.TypedName, properties px.OrderedMap) serviceapi.Definition {
	return &definition{identifier, serviceId, properties}
}

type definition struct {
	identifier px.TypedName
	serviceId  px.TypedName
	properties px.OrderedMap
}

func (d *definition) Label() string {
	return fmt.Sprintf(`%s/%s`, d.serviceId.Name(), d.identifier.Name())
}

func (d *definition) Get(key string) (value px.Value, ok bool) {
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

func (d *definition) InitHash() px.OrderedMap {
	es := make([]*types.HashEntry, 0, 3)
	es = append(es, types.WrapHashEntry2(`identifier`, d.identifier))
	es = append(es, types.WrapHashEntry2(`serviceId`, d.serviceId))
	es = append(es, types.WrapHashEntry2(`properties`, d.properties))
	return types.WrapHash(es)
}

func (d *definition) Equals(other interface{}, g px.Guard) bool {
	if o, ok := other.(*definition); ok {
		return d.identifier.Equals(o.identifier, g) && d.serviceId.Equals(o.serviceId, g) && d.properties.Equals(o.properties, g)
	}
	return false
}

func (d *definition) Identifier() px.TypedName {
	return d.identifier
}

func (d *definition) ServiceId() px.TypedName {
	return d.serviceId
}

func (d *definition) Properties() px.OrderedMap {
	return d.properties
}

func (d *definition) String() string {
	return px.ToString(d)
}

func (d *definition) ToString(bld io.Writer, format px.FormatContext, g px.RDetect) {
	types.ObjectToString(d, format, bld, g)
}

func (d *definition) PType() px.Type {
	return serviceapi.DefinitionMetaType
}
