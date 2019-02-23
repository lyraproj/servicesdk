package typegen

import (
	"bufio"
	"github.com/lyraproj/puppet-evaluator/eval"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type puppetGenerator struct{}

func (g *puppetGenerator) GenerateTypes(typeSet eval.TypeSet, directory string) {
	g.GenerateType(typeSet, directory)
}

func (g *puppetGenerator) GenerateType(typ eval.Type, directory string) {
	typeToStream(typ, directory, `.pp`, func(b io.Writer) {
		write(b, "# this file is generated\ntype ")
		write(b, typ.Name())
		write(b, " = ")
		typ.ToString(b, eval.PRETTY_EXPANDED, nil)
		write(b, "\n")
	})
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

func typeToStream(typ eval.Type, directory, extension string, gen func(io.Writer)) {
	tsp := strings.Split(typ.Name(), `::`)
	fn := filepath.Join(directory, filepath.Join(tsp...)) + extension

	err := os.MkdirAll(filepath.Dir(fn), os.ModePerm)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(fn)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b := bufio.NewWriter(f)
	gen(b)
	err = b.Flush()
	if err != nil {
		panic(err)
	}
}
