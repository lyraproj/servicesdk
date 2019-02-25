package lang

import "github.com/lyraproj/issue/issue"

const (
	WF_UNSUPPORTED_LANGUAGE = `WF_UNSUPPORTED_LANGUAGE`
)

func init() {
	issue.Hard(WF_UNSUPPORTED_LANGUAGE, `language %{language} not supported. Choose one of %{supportedLanguages}"`)
}
