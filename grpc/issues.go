package grpc

import "github.com/lyraproj/issue/issue"

const (
	InvocationError = `WF_INVOCATION_ERROR`
)

func init() {
	issue.Hard(InvocationError, `invocation of %{identifier} %{name} failed: %{code} %{message}`)
}
