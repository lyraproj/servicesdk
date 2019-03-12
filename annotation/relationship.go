package annotation

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

var RelationshipType px.ObjectType

const KindContained = `contained`
const KindContainer = `container`
const KindConsumer = `consumer`
const KindProvider = `provider`

const CardinalityOne = `one`
const CardinalityMany = `many`
const CardinalityZeroOrOne = `zeroOrOne`

type Relationship struct {
	Type        px.Type
	Kind        string   `puppet:"type => Enum[contained, container, consumer, provider]"`
	Cardinality string   `puppet:"type => Enum[one, many, zeroOrOne]"`
	Keys        []string `puppet:"type => Array[Pcore::MemberName]"`
	ReverseName *string  `puppet:"type => Pcore::MemberName, value => undef"`
}

func init() {
	RelationshipType = px.NewGoType(`Lyra::Relationship`, Relationship{})
}

func (r *Relationship) Validate(c px.Context, typ px.ObjectType, name string) {
	at, ok := r.Type.(px.ObjectType)
	if !ok {
		panic(px.Error(RelationshipTypeIsNotObject, issue.H{`type`: r.Type}))
	}

	nk := len(r.Keys)
	if nk%2 != 0 {
		panic(px.Error(RelationshipKeysUnevenNumber, issue.H{`type`: r.Type}))
	}

	for i := 0; i < nk; i += 2 {
		assertAttribute(typ, r.Keys[i])
		assertAttribute(at, r.Keys[i+1])
	}

	var rs Resource
	ra, ok := at.Annotations(c).Get(ResourceType)
	if ok {
		rs, ok = ra.(Resource)
	}
	if !ok {
		panic(px.Error(NoResourceAnnotation, issue.H{`type`: r.Type}))
	}

	var cr, v *Relationship
	cs := rs.Relationships()
	if r.ReverseName != nil {
		if v, ok = cs[*r.ReverseName]; ok && v.IsCounterpartOf(name, typ, r) {
			cr = v
		}
	} else {
		for _, v = range cs {
			if v.IsCounterpartOf(name, typ, r) {
				if cr != nil {
					panic(px.Error(MultipleCounterparts, issue.H{`type`: r.Type, `name`: name}))
				}
				cr = v
			}
		}
	}
	if cr == nil {
		panic(px.Error(CounterpartNotFound, issue.H{`type`: r.Type, `name`: name}))
	}
}

func (r *Relationship) IsCounterpartOf(name string, typ px.ObjectType, o *Relationship) (match bool) {
	switch r.Kind {
	case KindContained:
		match = o.Kind == KindContainer
	case KindContainer:
		match = o.Kind == KindContained
	case KindConsumer:
		match = o.Kind == KindProvider
	case KindProvider:
		match = o.Kind == KindConsumer
	default:
		match = false
	}

	if match {
		switch r.Cardinality {
		case CardinalityMany:
			match = o.Cardinality != CardinalityMany
		case CardinalityOne:
			match = o.Cardinality != CardinalityOne
		case CardinalityZeroOrOne:
		default:
			match = false
		}
	}

	if match && r.ReverseName != nil {
		match = name == *r.ReverseName
	}

	if match {
		nk := len(r.Keys)
		match = nk == len(o.Keys)
		if match {
			// Must match in reverse
			nk--
			for i, k := range r.Keys {
				if k != o.Keys[nk-i] {
					match = false
					break
				}
			}
		}
	}
	if match {
		match = r.Type.Equals(typ, nil)
	}
	return
}
