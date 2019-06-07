package grpc

import "github.com/lyraproj/issue/issue"

const (
	InvocationError       = `WF_INVOCATION_ERROR`
	ProcInvocationError   = `WF_PROC_INVOCATION_ERROR`
	RemoteInvocationError = `WF_REMOTE_INVOCATION_ERROR`
)

func init() {
	issue.Hard(RemoteInvocationError, `Failed to invoke method %{executable}#%{identifier}/%{name}() on host %{host}`)
	issue.Hard(ProcInvocationError, `Failed to invoke method %{executable}#%{identifier}/%{name}()`)
	issue.Hard(InvocationError, `Failed to invoke method %{identifier}/%{name}()`)
}
