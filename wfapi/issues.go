package wfapi

import "github.com/puppetlabs/go-issues/issue"

const (
	WF_ILLEGAL_ITERATION_STYLE   = `WF_ILLEGAL_ITERATION_STYLE`
	WF_ILLEGAL_OPERATION         = `WF_ILLEGAL_OPERATION`
	WF_ACTIVITY_NO_NAME          = `WF_ACTIVITY_NO_NAME`
	WF_ITERATOR_NOT_ONE_ACTIVITY = `WF_ITERATOR_NOT_ONE_ACTIVITY`
)

func init() {
	issue.Hard(WF_ILLEGAL_ITERATION_STYLE, `no such iteration style '%{style}'`)
	issue.Hard(WF_ILLEGAL_OPERATION, `no such operation '%{operation}'`)
	issue.Hard(WF_ACTIVITY_NO_NAME, `an activity must have a name`)
	issue.Hard(WF_ITERATOR_NOT_ONE_ACTIVITY, `an iterator must have exactly one activity`)
}
