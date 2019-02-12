package service

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/annotation"
	"reflect"
)

type ResourceTypeBuilder interface {
	AddRelationship(name, to, kind, cardinality, reverse_name string, keys []string)
	ImmutableAttributes(names ...string)
	ProvidedAttributes(names ...string)
	Tags(tags map[string]string)
	Build(goType interface{}) eval.AnnotatedType
}

type rtBuilder struct {
	ctx            eval.Context
	rels           []*types.HashEntry
	immutableAttrs []string
	providedAttrs  []string
	tags           map[string]string
}

func (rb *rtBuilder) AddRelationship(name, to, kind, cardinality, reverseName string, keys []string) {
	ln := 4
	if reverseName != `` {
		ln++
	}
	es := make([]*types.HashEntry, ln)
	es[0] = types.WrapHashEntry2(`type`, types.NewTypeReferenceType(to))
	es[1] = types.WrapHashEntry2(`kind`, types.WrapString(kind))
	es[2] = types.WrapHashEntry2(`cardinality`, types.WrapString(cardinality))
	es[3] = types.WrapHashEntry2(`keys`, types.WrapStrings(keys))
	if reverseName != `` {
		es[4] = types.WrapHashEntry2(`reverseName`, types.WrapString(reverseName))
	}
	rb.rels = append(rb.rels, types.WrapHashEntry2(name, types.WrapHash(es)))
}

func (rb *rtBuilder) ImmutableAttributes(names ...string) {
	if rb.immutableAttrs == nil {
		rb.immutableAttrs = names
	} else {
		rb.immutableAttrs = append(rb.immutableAttrs, names...)
	}
}

func (rb *rtBuilder) ProvidedAttributes(names ...string) {
	if rb.providedAttrs == nil {
		rb.providedAttrs = names
	} else {
		rb.providedAttrs = append(rb.providedAttrs, names...)
	}
}

func (rb *rtBuilder) Tags(tags map[string]string) {
	if rb.tags == nil {
		rb.tags = tags
	} else {
		for k, v := range tags {
			rb.tags[k] = v
		}
	}
}

func (rb *rtBuilder) Build(goType interface{}) eval.AnnotatedType {
	var rt reflect.Type
	switch goType.(type) {
	case reflect.Type:
		rt = goType.(reflect.Type)
	case reflect.Value:
		rt = goType.(reflect.Value).Type()
	default:
		rt = reflect.TypeOf(goType)
	}

	annotations := eval.EMPTY_MAP
	if rb.immutableAttrs != nil || rb.providedAttrs != nil || rb.rels != nil {
		as := make([]*types.HashEntry, 0, 3)
		if rb.immutableAttrs != nil {
			as = append(as, types.WrapHashEntry2(`immutableAttributes`, types.WrapStrings(rb.immutableAttrs)))
		}
		if rb.providedAttrs != nil {
			as = append(as, types.WrapHashEntry2(`providedAttributes`, types.WrapStrings(rb.providedAttrs)))
		}
		if rb.rels != nil {
			as = append(as, types.WrapHashEntry2(`relationships`, types.WrapHash(rb.rels)))
		}
		annotations = types.SingletonHash(annotation.ResourceType, types.WrapHash(as))
	}
	return eval.NewAnnotatedType(rt, rb.tags, annotations)
}
