package typegen

import (
	"io"

	"github.com/lyraproj/pcore/px"
)

type puppetGenerator struct{}

func (g *puppetGenerator) GenerateTypes(typeSet px.TypeSet, directory string) {
	g.GenerateType(typeSet, directory)
}

func (g *puppetGenerator) GenerateType(typ px.Type, directory string) {
	typeToStream(typeFile(typ, directory, `.pp`), func(b io.Writer) {
		write(b, "# this file is generated\ntype ")
		write(b, typ.Name())
		write(b, " = ")
		typ.ToString(b, px.PrettyExpanded, nil)
		write(b, "\n")
	})
}
