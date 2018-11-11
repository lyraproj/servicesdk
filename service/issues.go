package service

import "github.com/puppetlabs/go-issues/issue"

const (
	WF_ALREADY_REGISTERED = `WF_ALREADY_REGISTERED`
	WF_ILLEGAL_TYPE_NAME  = `WF_ILLEGAL_TYPE_NAME`
	WF_NO_SUCH_METHOD     = `WF_NO_SUCH_METHOD`
)

func init() {
	issue.Hard(WF_ALREADY_REGISTERED, `the %{namespace} %{identifier} API has already been registered`)
	issue.Hard(WF_ILLEGAL_TYPE_NAME, `name must be segments starting with an uppercase letter joined with'::'. Got: '%{name}'`)
	issue.Hard(WF_NO_SUCH_METHOD, `the '%{api}' API does not have a method named %{method}`)
}
