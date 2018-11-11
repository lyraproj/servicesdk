package wfapi

import "github.com/puppetlabs/go-issues/issue"

const (
	WF_ILLEGAL_ITERATION_STYLE = `WF_ILLEGAL_ITERATION_STYLE`
	WF_ILLEGAL_OPERATION       = `WF_ILLEGAL_OPERATION`
)

func init() {
	issue.Hard(WF_ILLEGAL_ITERATION_STYLE, `no such iteration style '%{style}'`)
	issue.Hard(WF_ILLEGAL_OPERATION, `no such operation '%{operation}'`)
}
