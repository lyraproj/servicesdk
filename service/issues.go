package service

import "github.com/lyraproj/issue/issue"

const (
	AlreadyRegistered    = `WF_ALREADY_REGISTERED`
	ApiTypeNotRegistered = `WF_API_TYPE_NOT_REGISTERED`
	IllegalTypeName      = `WF_ILLEGAL_TYPE_NAME`
	NoCommonNamespace    = `WF_NO_COMMON_NAMESPACE`
	NoSuchApi            = `WF_NO_SUCH_API`
	NoSuchMethod         = `WF_NO_SUCH_METHOD`
	NoSuchState          = `WF_NO_SUCH_STATE`
	NotFound             = `WF_NOT_FOUND`
	NotFunc              = `WF_NOT_FUNC`
	NotPuppetObject      = `WF_NOT_PUPPET_OBJECT`
	NoStateConverter     = `WF_NO_STATE_CONVERTER`
	TypeNameClash        = `WF_TYPE_NAME_CLASH`
)

func init() {
	issue.Hard(AlreadyRegistered, `the %{namespace} %{identifier} API has already been registered`)
	issue.Hard(ApiTypeNotRegistered, `the Go type %{type} has not been registered as an API type`)
	issue.Hard(IllegalTypeName, `name must be segments starting with an uppercase letter joined with'::'. Got: '%{name}'`)
	issue.Hard(NoCommonNamespace, `registered types share no common namespace`)
	issue.Hard(NoSuchApi, `the '%{api}' API does not exist`)
	issue.Hard(NoSuchMethod, `the '%{api}' API does not have a method named %{method}`)
	issue.Hard(NoSuchState, `state '%{name}' not found`)
	issue.Hard(NoStateConverter, `no state converter has been registered`)
	issue.Hard(NotFound, `%{typeName} resource with external id '%{extId}' does not exist`)
	issue.Hard(NotFunc, `attempt to register a function '%{name}' as a %{type}. Expected a func'`)
	issue.Hard(NotPuppetObject, `expected resource to produce an Object, got '%{actual}'`)
	issue.Hard(TypeNameClash, `attempt to register '%{goType}' using both '%{oldType}' and '%{newType}'`)
}
