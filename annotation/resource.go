package annotation

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"io"
	"reflect"
	"sort"
)

var ResourceType eval.ObjectType

func init() {
	ResourceType = eval.NewObjectType(`Lyra::Resource`, `Annotation{
    attributes => {
      # provided_attributes lists the names of the attributes that originates
      # from the cloud provider and shouldn't be used in comparison between
      # desired state an actual state.
      provided_attributes => Optional[Array[Pcore::MemberName]],

      # relationships describe how the annotated resource type relates to
      # other resource types.
      relationships => Optional[Hash[Pcore::MemberName, Init[Lyra::Relationship]]]
    }
  }`,

		func(ctx eval.Context, args []eval.Value) eval.Value {
			switch len(args) {
			case 0:
				return NewResource(ctx, nil, nil)
			case 1:
				return NewResource(ctx, args[0], nil)
			default:
				return NewResource(ctx, args[0], args[1])
			}
		},

		func(ctx eval.Context, args []eval.Value) eval.Value {
			h := args[0].(*types.HashValue)
			return NewResource(ctx, h.Get5(`provided_attributes`, eval.UNDEF), h.Get5(`relationships`, eval.UNDEF))
		})
}

type Resource interface {
	eval.PuppetObject

	ProvidedAttributes() []string

	Relationships() map[string]*Relationship
}

type resource struct {
	providedAttributes []string
	relationships      map[string]*Relationship
}

func NewResource(ctx eval.Context, provided_attributes eval.Value, relationships eval.Value) Resource {
	r := &resource{}
	if pa, ok := provided_attributes.(*types.ArrayValue); ok {
		r.providedAttributes = eval.StringElements(pa)
	}
	if rs, ok := relationships.(eval.OrderedMap); ok {
		rels := make(map[string]*Relationship, rs.Len())
		rs.EachPair(func(k, v eval.Value) {
			rv := eval.New(ctx, RelationshipType, v).(eval.Reflected).Reflect(ctx)
			rels[k.String()] = rv.Addr().Interface().(*Relationship)
		})
		r.relationships = rels
	}
	return r
}

func (r *resource) ProvidedAttributes() []string {
	return r.providedAttributes
}

func (r *resource) ProvidedAttributesList() eval.Value {
	if r.providedAttributes == nil {
		return eval.UNDEF
	}
	return types.WrapStrings(r.providedAttributes)
}

func (r *resource) Relationships() map[string]*Relationship {
	return r.relationships
}

func (r *resource) RelationshipsMap() eval.Value {
	if r.relationships == nil {
		return eval.UNDEF
	}
	es := make([]*types.HashEntry, len(r.relationships))
	for k, v := range r.relationships {
		es = append(es, types.WrapHashEntry2(k, types.NewReflectedValue(RelationshipType, reflect.ValueOf(v))))
	}
	// Sort by key to get predictable order
	sort.Slice(es, func(i, j int) bool { return es[i].Key().String() < es[j].Key().String() })
	return types.WrapHash(es)
}

func (r *resource) Validate(c eval.Context, annotatedType eval.Annotatable) {
	ot, ok := annotatedType.(eval.ObjectType)
	if !ok {
		panic(eval.Error(RA_ANNOTATED_IS_NOT_OBJECT, issue.H{`type`: annotatedType}))
	}
	if r.relationships != nil {
		isContained := false
		for k, v := range r.relationships {
			v.Validate(c, ot, k)
			if v.Kind == KindContained {
				if isContained {
					panic(eval.Error(RA_CONTAINED_MORE_THAN_ONCE, issue.H{`type`: ot}))
				}
				isContained = true
			}
		}
	}
	if r.providedAttributes != nil {
		for _, p := range r.providedAttributes {
			if m, ok := ot.Member(p); ok {
				if a, ok := m.(eval.Attribute); ok {
					if a.HasValue() {
						continue
					}
					panic(eval.Error(RA_PROVIDED_ATTRIBUTE_IS_REQUIRED, issue.H{`attr`: a}))
				}
			}
			panic(eval.Error(RA_PROVIDED_ATTRIBUTE_NOT_FOUND, issue.H{`type`: ot, `name`: p}))
		}
	}
}

func (r *resource) String() string {
	return eval.ToString(r)
}

func (r *resource) Equals(other interface{}, guard eval.Guard) bool {
	if or, ok := other.(*resource); ok {
		return eval.Equals(r.providedAttributes, or.providedAttributes) && eval.Equals(r.relationships, or.relationships)
	}
	return false
}

func (r *resource) ToString(bld io.Writer, format eval.FormatContext, g eval.RDetect) {
	types.ObjectToString(r, format, bld, g)
}

func (r *resource) PType() eval.Type {
	return ResourceType
}

func (r *resource) Get(key string) (value eval.Value, ok bool) {
	switch key {
	case `provided_attributes`:
		return r.ProvidedAttributesList(), true
	case `relationships`:
		return r.RelationshipsMap(), true
	}
	return nil, false
}

func (r *resource) InitHash() eval.OrderedMap {
	es := make([]*types.HashEntry, 3)
	if r.providedAttributes != nil {
		es = append(es, types.WrapHashEntry2(`provided_attributes`, r.ProvidedAttributesList()))
	}
	if r.relationships != nil {
		es = append(es, types.WrapHashEntry2(`relationships`, r.RelationshipsMap()))
	}
	return types.WrapHash(es)
}
