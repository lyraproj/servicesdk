package typegen

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/lang"
	"sort"
	"strings"
)

// The Generator interface is implemented by generators that can transform Pcore types
// to types in some specific language.
type Generator interface {
	// GenerateTypes produces types in some language for all types in the given TypeSet and writes
	// them to a file under the given directory
	GenerateTypes(ts eval.TypeSet, directory string)

	// GenerateType produces a type in some language and writes it to a file under the
	// given directory
	GenerateType(t eval.Type, directory string)
}

// All known language generators
var generators = map[string]Generator{
	"puppet":     &puppetGenerator{},
	"typescript": &tsGenerator{},
}

func GetGenerator(language string) Generator {
	generator, ok := generators[strings.ToLower(language)]
	if !ok {
		sl := make([]string, 0, len(generators))
		for l, _ := range generators {
			sl = append(sl, l)
		}
		sort.Strings(sl)
		panic(eval.Error(lang.WF_UNSUPPORTED_LANGUAGE,
			issue.H{`language`: language, `supportedLanguages`: strings.Join(sl, `, `)}))
	}
	return generator
}
