package service

import (
	"fmt"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
)

type subService struct {
	def serviceapi.Definition
}

func NewSubService(def serviceapi.Definition) serviceapi.Service {
	return &subService{def}
}

func (s *subService) Parent(c px.Context) serviceapi.Service {
	x, ok := px.Load(c, s.def.ServiceId())
	if !ok {
		panic(fmt.Errorf("failed to load %s", s.def.ServiceId()))
	}
	return x.(serviceapi.Service)
}

func (s *subService) Invoke(c px.Context, identifier, name string, arguments ...px.Value) px.Value {
	args := make([]px.Value, 2, 2+len(arguments))
	args[0] = types.WrapString(identifier)
	args[1] = types.WrapString(name)
	args = append(args, arguments...)
	return s.Parent(c).Invoke(c, s.def.Identifier().Name(), "invoke", args...)
}

func (s *subService) Metadata(c px.Context) (typeSet px.TypeSet, definitions []serviceapi.Definition) {
	v := s.Parent(c).Invoke(c, s.def.Identifier().Name(), "metadata").(px.List)
	if ts, ok := v.At(0).(px.TypeSet); ok {
		typeSet = ts
	}
	if dl, ok := v.At(1).(px.List); ok {
		definitions = make([]serviceapi.Definition, dl.Len())
		dl.EachWithIndex(func(d px.Value, i int) {
			definitions[i] = d.(serviceapi.Definition)
		})
	}
	return
}

func (s *subService) State(c px.Context, name string, input px.OrderedMap) px.PuppetObject {
	return s.Parent(c).Invoke(c, s.def.Identifier().Name(), "state", types.WrapString(name), input).(px.PuppetObject)
}

func (s *subService) Identifier(px.Context) px.TypedName {
	return px.NewTypedName(px.NsService, s.def.Identifier().Name())
}
