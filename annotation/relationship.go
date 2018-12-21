package annotation

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
)

var RelationshipType eval.ObjectType

const KindContained = `contained`
const KindContainer = `container`
const KindConsumer = `consumer`
const KindProvider = `provider`

type Relationship struct {
	Type        eval.Type
	Kind        string   `puppet:"type => Enum[contained, container, consumer, provider]"`
	Cardinality string   `puppet:"type => Enum[one, many, zero_or_one]"`
	Keys        []string `puppet:"type => Array[Pcore::MemberName]"`
	ReverseName *string  `puppet:"type => Pcore::MemberName, value => undef"`
}

func init() {
	RelationshipType = eval.NewGoType(`Lyra::Relationship`, &Relationship{})
}

func (r *Relationship) Validate(c eval.Context, typ eval.ObjectType, name string) {
	at, ok := r.Type.(eval.ObjectType)
	if !ok {
		panic(eval.Error(RA_RELATIONSHIP_TYPE_IS_NOT_OBJECT, issue.H{`type`: r.Type}))
	}
	var rs Resource
	ra, ok := at.Annotations().Get(ResourceType)
	if ok {
		rs, ok = ra.(Resource)
	}
	if !ok {
		panic(eval.Error(RA_NO_RESOURCE_ANNOTATION, issue.H{`type`: r.Type}))
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
					panic(eval.Error(RA_MULTIPLE_COUNTERPARTS, issue.H{`type`: r.Type, `name`: name}))
				}
				cr = v
			}
		}
	}
	if cr == nil {
		panic(eval.Error(RA_COUNTERPART_NOT_FOUND, issue.H{`type`: r.Type, `name`: name}))
	}
}

func (r *Relationship) IsCounterpartOf(name string, typ eval.ObjectType, o *Relationship) (match bool) {
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
	if match && r.ReverseName != nil {
		match = name == *r.ReverseName
	}
	if match {
		match = r.Type.Equals(typ, nil)
	}
	return
}
