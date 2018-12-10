package grpc

import "github.com/lyraproj/issue/issue"

const (
	WF_INVOCATION_ERROR = `WF_INVOCATION_ERROR`
)

func init() {
	issue.Hard(WF_INVOCATION_ERROR, `invocation of %{identifier} %{name} failed: %{code} %{message}`)
}
