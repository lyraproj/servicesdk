package service

import "github.com/puppetlabs/go-issues/issue"

const (
	WF_ALREADY_REGISTERED  = `WF_ALREADY_REGISTERED`
	WF_ILLEGAL_TYPE_NAME   = `WF_ILLEGAL_TYPE_NAME`
	WF_NO_COMMON_NAMESPACE = `WF_NO_COMMON_NAMESPACE`
	WF_NO_SUCH_METHOD      = `WF_NO_SUCH_METHOD`
	WF_NO_SUCH_STATE       = `WF_NO_SUCH_STATE`
	WF_NOT_PUPPET_OBJECT   = `WF_NOT_PUPPET_OBJECT`
	WF_NO_STATE_CONVERTER  = `WF_NO_STATE_CONVERTER`
)

func init() {
	issue.Hard(WF_ALREADY_REGISTERED, `the %{namespace} %{identifier} API has already been registered`)
	issue.Hard(WF_ILLEGAL_TYPE_NAME, `name must be segments starting with an uppercase letter joined with'::'. Got: '%{name}'`)
	issue.Hard(WF_NO_COMMON_NAMESPACE, `registered types share no common namespace`)
	issue.Hard(WF_NO_SUCH_METHOD, `the '%{api}' API does not have a method named %{method}`)
	issue.Hard(WF_NO_SUCH_STATE, `state '%{name}' not found`)
	issue.Hard(WF_NO_STATE_CONVERTER, `no state converter has been registered`)
	issue.Hard(WF_NOT_PUPPET_OBJECT, `expected resource to produce an Object, got '%{actual}'`)
}
