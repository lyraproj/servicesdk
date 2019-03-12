package annotation

import (
	"io"
	"reflect"
	"sort"

	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/pcore/utils"
)

var ResourceType px.ObjectType

func init() {
	ResourceType = px.NewGoObjectType(`Lyra::Resource`, reflect.TypeOf((*Resource)(nil)).Elem(), `Annotation{
    attributes => {
      # immutableAttributes lists the names of the attributes that cannot be
      # changed. If a difference is detected between the desired state and the
      # actual state that involves immutable attributes, then the resource must
      # be deleted and recreated in order to reach the desired state.
      immutableAttributes => Optional[Array[Pcore::MemberName]],

      # providedAttributes lists the names of the attributes that originates
      # from the resource provider and shouldn't be used in comparison between
      # desired state an actual state.
      providedAttributes => Optional[Array[Pcore::MemberName]],

      # relationships describe how the annotated resource type relates to
      # other resource types.
      relationships => Optional[Hash[Pcore::MemberName, Init[Lyra::Relationship]]]
    }
  }`,

		func(ctx px.Context, args []px.Value) px.Value {
			switch len(args) {
			case 0:
				return NewResource(ctx, nil, nil, nil)
			case 1:
				return NewResource(ctx, args[0], nil, nil)
			case 2:
				return NewResource(ctx, args[0], args[1], nil)
			default:
				return NewResource(ctx, args[0], args[1], args[3])
			}
		},

		func(ctx px.Context, args []px.Value) px.Value {
			h := args[0].(*types.Hash)
			return NewResource(ctx, h.Get5(`immutableAttributes`, px.Undef), h.Get5(`providedAttributes`, px.Undef), h.Get5(`relationships`, px.Undef))
		})
}

type Resource interface {
	px.PuppetObject

	// Changed returns two booleans.
	//
	// The first boolean is true when the value of an attribute differs between the desired and actual
	// state. Attributes listed in the providedAttributes array, for which the desired value is the
	// default, are exempt from the comparison.
	//
	// The second boolean is true when the first is true and the attribute in question is listed in the
	// immutableAttributes array.
	Changed(x, y px.PuppetObject) (bool, bool)

	ImmutableAttributes() []string

	ProvidedAttributes() []string

	Relationships() map[string]*Relationship
}

type resource struct {
	immutableAttributes []string
	providedAttributes  []string
	relationships       map[string]*Relationship
}

func NewResource(ctx px.Context, immutableAttributes, providedAttributes px.Value, relationships px.Value) Resource {
	r := &resource{}

	stringsOrNil := func(v px.Value) []string {
		if a, ok := v.(*types.Array); ok {
			sa := px.StringElements(a)
			if len(sa) > 0 {
				return sa
			}
		}
		return nil
	}

	r.immutableAttributes = stringsOrNil(immutableAttributes)
	r.providedAttributes = stringsOrNil(providedAttributes)
	if rs, ok := relationships.(px.OrderedMap); ok {
		rls := make(map[string]*Relationship, rs.Len())
		rs.EachPair(func(k, v px.Value) {
			rv := px.New(ctx, RelationshipType, v).(px.Reflected).Reflect(ctx)
			rls[k.String()] = rv.Addr().Interface().(*Relationship)
		})
		r.relationships = rls
	}
	return r
}

func (r *resource) ImmutableAttributes() []string {
	return r.immutableAttributes
}

func (r *resource) ImmutableAttributesList() px.Value {
	if r.immutableAttributes == nil {
		return px.Undef
	}
	return types.WrapStrings(r.immutableAttributes)
}

func (r *resource) ProvidedAttributes() []string {
	return r.providedAttributes
}

func (r *resource) ProvidedAttributesList() px.Value {
	if r.providedAttributes == nil {
		return px.Undef
	}
	return types.WrapStrings(r.providedAttributes)
}

func (r *resource) Relationships() map[string]*Relationship {
	return r.relationships
}

func (r *resource) RelationshipsMap() px.Value {
	if r.relationships == nil {
		return px.Undef
	}
	es := make([]*types.HashEntry, len(r.relationships))
	for k, v := range r.relationships {
		es = append(es, types.WrapHashEntry2(k, types.NewReflectedValue(RelationshipType, reflect.ValueOf(v))))
	}
	// Sort by key to get predictable order
	sort.Slice(es, func(i, j int) bool { return es[i].Key().String() < es[j].Key().String() })
	return types.WrapHash(es)
}

func (r *resource) Validate(c px.Context, annotatedType px.Annotatable) {
	ot, ok := annotatedType.(px.ObjectType)
	if !ok {
		panic(px.Error(AnnotatedIsNotObject, issue.H{`type`: annotatedType}))
	}
	if r.relationships != nil {
		isContained := false
		for k, v := range r.relationships {
			v.Validate(c, ot, k)
			if v.Kind == KindContained {
				if isContained {
					panic(px.Error(ContainedMoreThanOnce, issue.H{`type`: ot}))
				}
				isContained = true
			}
		}
	}
	if r.immutableAttributes != nil {
		for _, p := range r.immutableAttributes {
			assertAttribute(ot, p)
		}
	}
	if r.providedAttributes != nil {
		for _, p := range r.providedAttributes {
			a := assertAttribute(ot, p)
			if a.HasValue() {
				continue
			}
			panic(px.Error(ProvidedAttributeIsRequired, issue.H{`attr`: a}))
		}
	}
}

// Changed returns two booleans.
//
// The first boolean is true when the value of an attribute differs between the desired and actual
// state. Attributes listed in the providedAttributes array, for which the desired value is the
// default, are exempt from the comparison.
//
// The second boolean is true when the first is true and the attribute in question is listed in the
// immutableAttributes array.
func (r *resource) Changed(desired, actual px.PuppetObject) (bool, bool) {
	typ := desired.PType().(px.ObjectType)
	for _, a := range typ.AttributesInfo().Attributes() {
		dv := a.Get(desired)
		if r.isProvided(a.Name()) && a.Default(dv) {
			continue
		}
		av := a.Get(actual)
		if !dv.Equals(av, nil) {
			log := hclog.Default()
			if r.isImmutable(a.Name()) {
				log.Debug("immutable attribute mismatch", "attribute", a.Label(), "desired", dv, "actual", av)
				return true, true
			}
			log.Debug("mutable attribute mismatch", "attribute", a.Label(), "desired", dv, "actual", av)
			return true, false
		}
	}
	return false, false
}

func (r *resource) String() string {
	return px.ToString(r)
}

func (r *resource) Equals(other interface{}, guard px.Guard) bool {
	if or, ok := other.(*resource); ok {
		return px.Equals(r.providedAttributes, or.providedAttributes, guard) && px.Equals(r.relationships, or.relationships, guard)
	}
	return false
}

func (r *resource) ToString(bld io.Writer, format px.FormatContext, g px.RDetect) {
	types.ObjectToString(r, format, bld, g)
}

func (r *resource) PType() px.Type {
	return ResourceType
}

func (r *resource) Get(key string) (value px.Value, ok bool) {
	switch key {
	case `immutableAttributes`:
		return r.ImmutableAttributesList(), true
	case `providedAttributes`:
		return r.ProvidedAttributesList(), true
	case `relationships`:
		return r.RelationshipsMap(), true
	}
	return nil, false
}

func (r *resource) InitHash() px.OrderedMap {
	es := make([]*types.HashEntry, 3)
	if r.immutableAttributes != nil {
		es = append(es, types.WrapHashEntry2(`immutableAttributes`, r.ImmutableAttributesList()))
	}
	if r.providedAttributes != nil {
		es = append(es, types.WrapHashEntry2(`providedAttributes`, r.ProvidedAttributesList()))
	}
	if r.relationships != nil {
		es = append(es, types.WrapHashEntry2(`relationships`, r.RelationshipsMap()))
	}
	return types.WrapHash(es)
}

func assertAttribute(ot px.ObjectType, n string) (a px.Attribute) {
	if m, ok := ot.Member(n); ok {
		if a, ok = m.(px.Attribute); ok {
			return
		}
	}
	panic(px.Error(AttributeNotFound, issue.H{`type`: ot, `name`: n}))
}

func (r *resource) isProvided(name string) bool {
	return r.providedAttributes != nil && utils.ContainsString(r.providedAttributes, name)
}

func (r *resource) isImmutable(name string) bool {
	return r.immutableAttributes != nil && utils.ContainsString(r.immutableAttributes, name)
}
