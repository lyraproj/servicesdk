package lang

import "github.com/lyraproj/issue/issue"

const (
	UnsupportedLanguage = `WF_UNSUPPORTED_LANGUAGE`
)

func init() {
	issue.Hard(UnsupportedLanguage, `language %{language} not supported. Choose one of %{supportedLanguages}"`)
}
