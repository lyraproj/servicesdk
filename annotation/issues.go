package annotation

import "github.com/lyraproj/issue/issue"

const (
	RA_ANNOTATED_IS_NOT_OBJECT         = `RA_ANNOTATED_IS_NOT_OBJECT`
	RA_PROVIDED_ATTRIBUTE_IS_REQUIRED  = `RA_PROVIDED_ATTRIBUTE_IS_REQUIRED`
	RA_PROVIDED_ATTRIBUTE_NOT_FOUND    = `RA_PROVIDED_ATTRIBUTE_NOT_FOUND`
	RA_RELATIONSHIP_TYPE_IS_NOT_OBJECT = `RA_RELATIONSHIP_TYPE_IS_NOT_OBJECT`
	RA_NO_RESOURCE_ANNOTATION          = `RA_NO_RESOURCE_ANNOTATION`
	RA_MULTIPLE_COUNTERPARTS           = `RA_MULTIPLE_COUNTERPARTS`
	RA_COUNTERPART_NOT_FOUND           = `RA_COUNTERPART_NOT_FOUND`
	RA_CONTAINED_MORE_THAN_ONCE        = `RA_CONTAINED_MORE_THAN_ONCE`
)

func init() {
	issue.Hard2(RA_ANNOTATED_IS_NOT_OBJECT, `annotated %{type} is not an Object`, issue.HF{`attr`: issue.Label})
	issue.Hard2(RA_PROVIDED_ATTRIBUTE_IS_REQUIRED, `provided attribute %{attr} cannot be required`, issue.HF{`attr`: issue.Label})
	issue.Hard2(RA_PROVIDED_ATTRIBUTE_NOT_FOUND, `%{type} has no attribute named %{name}`, issue.HF{`type`: issue.Label})
	issue.Hard2(RA_RELATIONSHIP_TYPE_IS_NOT_OBJECT, `relationship type %{type} is not an Object`, issue.HF{`type`: issue.Label})
	issue.Hard2(RA_NO_RESOURCE_ANNOTATION, `relationship type %{type} has no Resource annotation`, issue.HF{`type`: issue.Label})
	issue.Hard2(RA_MULTIPLE_COUNTERPARTS, `relationship type %{type} has multiple matching counterparts for relation %{name}`, issue.HF{`type`: issue.Label})
	issue.Hard2(RA_COUNTERPART_NOT_FOUND, `relationship type %{type} has no matching counterpart for relation %{name}`, issue.HF{`type`: issue.Label})
	issue.Hard2(RA_CONTAINED_MORE_THAN_ONCE, `the type %{type} has more than one relationship of kind 'contained'`, issue.HF{`type`: issue.Label})
}
