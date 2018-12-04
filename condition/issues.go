package condition

import "github.com/lyraproj/issue/issue"

const (
	WF_CONDITION_SYNTAX_ERROR   = `WF_CONDITION_SYNTAX_ERROR`
	WF_CONDITION_MISSING_RP     = `WF_CONDITION_MISSING_RP`
	WF_CONDITION_INVALID_NAME   = `WF_CONDITION_INVALID_NAME`
	WF_CONDITION_UNEXPECTED_END = `WF_CONDITION_UNEXPECTED_END`
)

func init() {
	issue.Hard(WF_CONDITION_SYNTAX_ERROR, `syntax error in condition '%{text}' at position %{pos}`)
	issue.Hard(WF_CONDITION_MISSING_RP, `expected right parenthesis in condition '%{text}' at position %{pos}`)
	issue.Hard(WF_CONDITION_INVALID_NAME, `invalid name '%{name}' in condition '%{text}' at position %{pos}`)
	issue.Hard(WF_CONDITION_UNEXPECTED_END, `unexpected end of condition '%{text}' at position %{pos}`)
}
