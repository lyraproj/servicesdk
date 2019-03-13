package annotation

import "github.com/lyraproj/issue/issue"

const (
	AnnotatedIsNotObject         = `RA_ANNOTATED_IS_NOT_OBJECT`
	AttributeNotFound            = `RA_ATTRIBUTE_NOT_FOUND`
	ProvidedAttributeIsRequired  = `RA_PROVIDED_ATTRIBUTE_IS_REQUIRED`
	RelationshipKeysUnevenNumber = `RA_RELATIONSHIP_KEYS_UNEVEN_NUMBER`
	RelationshipTypeIsNotObject  = `RA_RELATIONSHIP_TYPE_IS_NOT_OBJECT`
	NoResourceAnnotation         = `RA_NO_RESOURCE_ANNOTATION`
	MultipleCounterparts         = `RA_MULTIPLE_COUNTERPARTS`
	CounterpartNotFound          = `RA_COUNTERPART_NOT_FOUND`
	ContainedMoreThanOnce        = `RA_CONTAINED_MORE_THAN_ONCE`
)

func init() {
	issue.Hard2(AnnotatedIsNotObject, `annotated %{type} is not an Object`, issue.HF{`attr`: issue.Label})
	issue.Hard2(AttributeNotFound, `%{type} has no attribute named %{name}`, issue.HF{`type`: issue.Label})
	issue.Hard2(ProvidedAttributeIsRequired, `provided attribute %{attr} cannot be required`, issue.HF{`attr`: issue.Label})
	issue.Hard2(RelationshipKeysUnevenNumber, `relationship type %{type} has an uneven number of keys`, issue.HF{`type`: issue.Label})
	issue.Hard2(RelationshipTypeIsNotObject, `relationship type %{type} is not an Object`, issue.HF{`type`: issue.Label})
	issue.Hard2(NoResourceAnnotation, `relationship type %{type} has no Resource annotation`, issue.HF{`type`: issue.Label})
	issue.Hard2(MultipleCounterparts, `relationship type %{type} has multiple matching counterparts for relation %{name}`, issue.HF{`type`: issue.Label})
	issue.Hard2(CounterpartNotFound, `relationship type %{type} has no matching counterpart for relation %{name}`, issue.HF{`type`: issue.Label})
	issue.Hard2(ContainedMoreThanOnce, `the type %{type} has more than one relationship of kind 'contained'`, issue.HF{`type`: issue.Label})
}
