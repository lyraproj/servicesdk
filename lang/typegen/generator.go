package typegen

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/lang"
)

// The Generator interface is implemented by generators that can transform Pcore types
// to types in some specific language.
type Generator interface {
	// GenerateTypes produces types in some language for all types in the given TypeSet and writes
	// them to a file under the given directory
	GenerateTypes(ts px.TypeSet, directory string)

	// GenerateType produces a type in some language and writes it to a file under the
	// given directory
	GenerateType(t px.Type, directory string)
}

// All known language generators
var generators = map[string]Generator{
	"go":         &goGeneratorFactory{},
	"puppet":     &puppetGenerator{},
	"typescript": &tsGeneratorFactory{},
}

func GetGenerator(language string) Generator {
	generator, ok := generators[strings.ToLower(language)]
	if !ok {
		sl := make([]string, 0, len(generators))
		for l := range generators {
			sl = append(sl, l)
		}
		sort.Strings(sl)
		panic(px.Error(lang.UnsupportedLanguage,
			issue.H{`language`: language, `supportedLanguages`: strings.Join(sl, `, `)}))
	}
	return generator
}

func writeByte(w io.Writer, b byte) {
	_, err := w.Write([]byte{b})
	if err != nil {
		panic(err)
	}
}

func write(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	if err != nil {
		panic(err)
	}
}

func typeFile(typ px.Type, directory, extension string) string {
	tsp := strings.Split(typ.Name(), `::`)
	return filepath.Join(directory, filepath.Join(tsp...)) + extension
}

func typeToStream(directory string, gen func(io.Writer)) {
	err := os.MkdirAll(filepath.Dir(directory), os.ModePerm)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(directory)
	if err != nil {
		panic(err)
	}
	//noinspection ALL
	defer f.Close()

	b := bufio.NewWriter(f)
	gen(b)
	err = b.Flush()
	if err != nil {
		panic(err)
	}
}
