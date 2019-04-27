package lyra

import (
	"github.com/lyraproj/issue/issue"
)

const (
	BadFunction             = `WF_BAD_FUNCTION`
	MissingRequiredField    = `WF_MISSING_ACTIVITY_NAME`
	MutuallyExclusiveFields = `WF_MUTUALLY_EXCLUSIVE_FIELDS`
	NotActionFunction       = `WF_NOT_STATE_FUNCTION`
	NotOneStructField       = `WF_NOT_ONE_STRUCT_FIELD`
	NotStateFunction        = `WF_NOT_STATE_FUNCTION`
	NotStruct               = `WF_NOT_STRUCT`
	RequireOneOfFields      = `WF_REQUIRE_ONE_OF_FIELDS`
)

func init() {
	issue.Hard(BadFunction, `the go func %{name} has invalid signature: %{type}`)
	issue.Hard(MissingRequiredField, `missing required field %{type}.%{name}`)
	issue.Hard2(MutuallyExclusiveFields, `only one of the %{fields} can have a value`, issue.HF{`fields`: issue.JoinErrors})
	issue.Hard(NotOneStructField, `struct describing parameter must have exactly one field, got %{type}`)
	issue.Hard(NotActionFunction, `expected action %{name} function to be a go func, got %{type}`)
	issue.Hard(NotStateFunction, `expected resource %{name} state function to be a go func, got %{type}`)
	issue.Hard(NotStruct, `%{name} argument must be a go struct or a pointer to a go struct, got '%{type}'`)
	issue.Hard2(RequireOneOfFields, `one of the %{fields} must have a value`, issue.HF{`fields`: issue.JoinErrors})
}
