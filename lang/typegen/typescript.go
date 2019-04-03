package typegen

import (
	"bytes"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/pcore/utils"
)

type tsGeneratorFactory struct {
}

type tsGenerator struct {
	anonIfds map[string]string
	ns       []string
	useIfds  bool
}

func (gf *tsGeneratorFactory) GenerateTypes(typeSet px.TypeSet, directory string) {
	tss := make([]px.TypeSet, 0)
	tts := make([]px.Type, 0)
	typeSet.Types().EachValue(func(t px.Value) {
		if ts, ok := t.(px.TypeSet); ok {
			tss = append(tss, ts)
		} else {
			tts = append(tts, t.(px.Type))
		}
	})

	for _, ts := range tss {
		gf.GenerateTypes(ts, directory)
	}

	if len(tts) > 0 {
		typeToStream(typeSet, directory, `.ts`, func(b io.Writer) {
			write(b, "// this file is generated\n")
			write(b, "import {PcoreValue, Value} from 'lyra-workflow';\n")
			g := &tsGenerator{make(map[string]string), strings.Split(typeSet.Name(), `::`), true}
			for _, t := range tts {
				g.generateType(t, 0, b)
			}
			g.writeAnonIfds(b)
		})
	}
}

func (gf *tsGeneratorFactory) GenerateType(typ px.Type, directory string) {
	typeToStream(typ, directory, `.ts`, func(b io.Writer) {
		write(b, "// this file is generated\n")
		write(b, "import {PcoreValue, Value} from 'lyra-workflow';\n")
		g := &tsGenerator{make(map[string]string), namespace(typ.Name()), true}
		g.generateType(typ, 0, b)
		g.writeAnonIfds(b)
	})
}

func (g *tsGenerator) writeAnonIfds(b io.Writer) {
	ac := len(g.anonIfds)
	if !g.useIfds || ac == 0 {
		return
	}

	names := make([]string, 0, ac)
	rev := make(map[string]string, ac)
	for sign, name := range g.anonIfds {
		names = append(names, name)
		rev[name] = sign
	}
	sort.Slice(names, func(i, j int) bool {
		ii, _ := strconv.Atoi(names[i][4:])
		ij, _ := strconv.Atoi(names[j][4:])
		return ii < ij
	})
	for _, name := range names {
		write(b, "\ninterface ")
		write(b, name)
		write(b, ` `)
		write(b, rev[name])
	}
}

// GenerateType produces a TypeScript type for the given Type and appends it to
// the given buffer.
func (g *tsGenerator) generateType(t px.Type, indent int, bld io.Writer) {
	if _, ok := t.(px.TypeSet); ok {
		return
	}

	if pt, ok := t.(px.ObjectType); ok {
		newLine(indent, bld)
		write(bld, `export class `)
		write(bld, nsName(g.ns, pt.Name()))
		if ppt, ok := pt.Parent().(px.ObjectType); ok {
			write(bld, ` extends `)
			write(bld, nsName(g.ns, ppt.Name()))
		} else {
			write(bld, ` implements PcoreValue`)
		}
		write(bld, ` {`)
		indent += 2
		ai := pt.AttributesInfo()
		allAttrs, thisAttrs, superAttrs := g.toTsAttrs(pt, ai.Attributes(), indent)
		g.appendFields(thisAttrs, indent, bld)
		if len(thisAttrs) > 0 {
			writeByte(bld, '\n')
		}
		if len(allAttrs) > 0 {
			g.appendConstructor(allAttrs, thisAttrs, superAttrs, indent, bld)
			writeByte(bld, '\n')
		}
		hasSuper := len(superAttrs) > 0
		if len(thisAttrs) > 0 || !hasSuper {
			g.appendPValueGetter(hasSuper, thisAttrs, indent, bld)
			writeByte(bld, '\n')
		}
		g.appendPTypeGetter(pt.Name(), indent, bld)
		indent -= 2
		newLine(indent, bld)
		write(bld, "}\n")
	} else {
		g.appendTsType(t, indent, bld)
	}
}

// ToTsType converts the given pType to a string representation of a TypeScript type. The given
// pType can not be a TypeSet.
func (g *tsGenerator) ToTsType(pType px.Type) string {
	return g.toTsType(pType, 0)
}

func (g *tsGenerator) toTsType(pType px.Type, indent int) string {
	bld := bytes.NewBufferString(``)
	g.appendTsType(pType, indent, bld)
	return bld.String()
}

type tsAttribute struct {
	tsName string
	name   string
	typ    string
	value  *string
}

var keywords = map[string]bool{
	// The following keywords are reserved and cannot be used as an Identifier:
	`arguments`:  true,
	`break`:      true,
	`case`:       true,
	`catch`:      true,
	`class`:      true,
	`const`:      true,
	`continue`:   true,
	`debugger`:   true,
	`default`:    true,
	`delete`:     true,
	`do`:         true,
	`else`:       true,
	`enum`:       true,
	`export`:     true,
	`extends`:    true,
	`false`:      true,
	`finally`:    true,
	`for`:        true,
	`function`:   true,
	`if`:         true,
	`import`:     true,
	`in`:         true,
	`instanceof`: true,
	`new`:        true,
	`null`:       true,
	`return`:     true,
	`super`:      true,
	`switch`:     true,
	`this`:       true,
	`throw`:      true,
	`true`:       true,
	`try`:        true,
	`typeof`:     true,
	`var`:        true,
	`void`:       true,
	`while`:      true,
	`with`:       true,

	// The following keywords cannot be used as identifiers in strict mode code, but are otherwise not restricted:
	`implements`: true,
	`interface`:  true,
	`let`:        true,
	`package`:    true,
	`private`:    true,
	`protected`:  true,
	`public`:     true,
	`static`:     true,
	`yield`:      true,

	// The following keywords cannot be used as user defined type names, but are otherwise not restricted:
	`any`:     true,
	`boolean`: true,
	`number`:  true,
	`string`:  true,
	`symbol`:  true,
}

func (g *tsGenerator) toTsAttrs(
	t px.ObjectType, attrs []px.Attribute, indent int) (allAttrs, thisAttrs, superAttrs []*tsAttribute) {
	allAttrs = make([]*tsAttribute, len(attrs))
	superAttrs = make([]*tsAttribute, 0)
	thisAttrs = make([]*tsAttribute, 0)
	for i, attr := range attrs {
		n := attr.Name()
		tsn := n
		if keywords[n] {
			tsn = n + `_`
		}
		tsAttr := &tsAttribute{tsName: tsn, name: n, typ: g.toTsType(attr.Type(), indent)}
		if attr.HasValue() {
			tsAttr.value = g.toTsValue(attr.Value())
		}
		if attr.Container() == t {
			thisAttrs = append(thisAttrs, tsAttr)
		} else {
			superAttrs = append(superAttrs, tsAttr)
		}
		allAttrs[i] = tsAttr
	}
	return
}

func (g *tsGenerator) appendFields(thisAttrs []*tsAttribute, indent int, bld io.Writer) {
	for _, attr := range thisAttrs {
		newLine(indent, bld)
		write(bld, `readonly `)
		write(bld, attr.tsName)
		write(bld, `: `)
		write(bld, attr.typ)
		write(bld, `;`)
	}
}

func (g *tsGenerator) appendConstructor(allAttrs, thisAttrs, superAttrs []*tsAttribute, indent int, bld io.Writer) {
	newLine(indent, bld)
	write(bld, `constructor(`)
	g.appendParameters(allAttrs, indent, bld)
	write(bld, `) {`)
	indent += 2
	if len(superAttrs) > 0 {
		newLine(indent, bld)
		write(bld, `super({`)
		for i, attr := range superAttrs {
			if i > 0 {
				write(bld, `, `)
			}
			write(bld, attr.tsName)
			write(bld, `: `)
			write(bld, attr.tsName)
		}
		write(bld, `});`)
	}
	for _, attr := range thisAttrs {
		newLine(indent, bld)
		write(bld, `this.`)
		write(bld, attr.tsName)
		write(bld, ` = `)
		write(bld, attr.tsName)
		writeByte(bld, ';')
	}
	indent -= 2
	newLine(indent, bld)
	writeByte(bld, '}')
}

func (g *tsGenerator) appendPValueGetter(hasSuper bool, thisAttrs []*tsAttribute, indent int, bld io.Writer) {
	newLine(indent, bld)
	write(bld, `__pvalue(): {[s: string]: Value} {`)
	indent += 2
	newLine(indent, bld)
	if len(thisAttrs) == 0 {
		if hasSuper {
			write(bld, `return super.__pvalue();`)
		} else {
			write(bld, `return {};`)
		}
	} else {
		if hasSuper {
			write(bld, `const ih = super.__pvalue();`)
		} else {
			write(bld, `const ih: {[s: string]: Value} = {};`)
		}
		for _, attr := range thisAttrs {
			newLine(indent, bld)
			if attr.value != nil {
				write(bld, `if (this.`)
				write(bld, attr.tsName)
				write(bld, ` !== `)
				write(bld, *attr.value)
				write(bld, `) {`)
				indent += 2
				newLine(indent, bld)
			}
			write(bld, `ih['`)
			write(bld, attr.name)
			write(bld, `'] = this.`)
			write(bld, attr.tsName)
			write(bld, `;`)
			if attr.value != nil {
				indent -= 2
				newLine(indent, bld)
				write(bld, `}`)
			}
		}
		newLine(indent, bld)
		write(bld, `return ih;`)
	}
	indent -= 2
	newLine(indent, bld)
	writeByte(bld, '}')
}

func (g *tsGenerator) appendPTypeGetter(name string, indent int, bld io.Writer) {
	newLine(indent, bld)
	write(bld, `__ptype(): string {`)
	indent += 2
	newLine(indent, bld)
	write(bld, `return '`)
	write(bld, name)
	write(bld, `';`)
	indent -= 2
	newLine(indent, bld)
	writeByte(bld, '}')
}

func (g *tsGenerator) appendParameters(params []*tsAttribute, indent int, bld io.Writer) {
	indent += 2
	write(bld, `{`)
	last := len(params) - 1
	for i, attr := range params {
		newLine(indent, bld)
		write(bld, attr.tsName)
		if attr.value != nil {
			write(bld, ` = `)
			write(bld, *attr.value)
		}
		if i < last {
			write(bld, `,`)
		}
	}
	indent -= 2
	newLine(indent, bld)
	write(bld, `}: {`)
	indent += 2

	for i, attr := range params {
		newLine(indent, bld)
		write(bld, attr.tsName)
		if attr.value != nil {
			writeByte(bld, '?')
		}
		write(bld, `: `)
		write(bld, attr.typ)
		if i < last {
			writeByte(bld, ',')
		}
	}

	indent -= 2
	newLine(indent, bld)
	write(bld, `}`)
}

func (g *tsGenerator) toTsValue(value px.Value) *string {
	bld := bytes.NewBufferString(``)
	g.appendTsValue(value, bld)
	s := bld.String()
	return &s
}

func (g *tsGenerator) appendTsValue(value px.Value, bld io.Writer) {
	switch value := value.(type) {
	case *types.UndefValue:
		write(bld, `null`)
	case px.StringValue:
		utils.PuppetQuote(bld, value.String())
	case px.Boolean, px.Integer, px.Float:
		write(bld, value.String())
	case *types.Array:
		writeByte(bld, '[')
		value.EachWithIndex(func(e px.Value, i int) {
			if i > 0 {
				write(bld, `, `)
			}
			g.appendTsValue(e, bld)
		})
		writeByte(bld, ']')
	case *types.Hash:
		writeByte(bld, '{')
		value.EachWithIndex(func(e px.Value, i int) {
			ev := e.(*types.HashEntry)
			if i > 0 {
				write(bld, `, `)
			}
			utils.PuppetQuote(bld, ev.Key().String())
			write(bld, `: `)
			g.appendTsValue(ev.Value(), bld)
		})
		writeByte(bld, '}')
	}
}

func (g *tsGenerator) appendTsType(pType px.Type, indent int, bld io.Writer) {
	switch pType := pType.(type) {
	case *types.BooleanType:
		write(bld, `boolean`)
	case *types.IntegerType, *types.FloatType:
		write(bld, `number`)
	case px.StringType:
		write(bld, `string`)
	case *types.OptionalType:
		g.appendTsType(pType.ContainedType(), indent, bld)
		write(bld, `|null`)
	case *types.ArrayType:
		et := pType.ElementType()
		switch et.(type) {
		case *types.ArrayType, *types.EnumType, *types.HashType, *types.OptionalType, *types.VariantType:
			write(bld, `Array<`)
			g.appendTsType(et, indent, bld)
			write(bld, `>`)
		default:
			g.appendTsType(et, indent, bld)
			write(bld, `[]`)
		}
	case *types.VariantType:
		for i, v := range pType.Types() {
			if i > 0 {
				write(bld, `|`)
			}
			g.appendTsType(v, indent, bld)
		}
	case *types.HashType:
		write(bld, `{[s: `)
		g.appendTsType(pType.KeyType(), indent, bld)
		write(bld, `]: `)
		g.appendTsType(pType.ValueType(), indent, bld)
		write(bld, `}`)
	case *types.EnumType:
		for i, s := range pType.Parameters() {
			if i > 0 {
				write(bld, `|`)
			}
			g.appendTsValue(s, bld)
		}
	case *types.TypeAliasType:
		write(bld, nsName(g.ns, pType.Name()))
	case px.ObjectType:
		if pType.Name() == `` {
			write(bld, g.makeAnonymousType(pType, indent))
		} else {
			write(bld, nsName(g.ns, pType.Name()))
		}
	}
}

func (g *tsGenerator) makeAnonymousType(t px.ObjectType, indent int) string {
	bld := bytes.NewBufferString(``)
	if g.useIfds {
		indent = 0
	}
	write(bld, `{`)
	indent += 2
	allAttrs, _, _ := g.toTsAttrs(t, t.AttributesInfo().Attributes(), indent)
	for i, a := range allAttrs {
		if i > 0 {
			write(bld, `,`)
		}
		newLine(indent, bld)
		write(bld, a.tsName)
		if a.value != nil {
			write(bld, `?`)
		}
		write(bld, `: `)
		write(bld, a.typ)
	}
	indent -= 2
	newLine(indent, bld)
	write(bld, `}`)
	sign := bld.String()
	if !g.useIfds {
		return sign
	}
	if prev, ok := g.anonIfds[sign]; ok {
		return prev
	}
	n := `Anon` + strconv.Itoa(len(g.anonIfds))
	g.anonIfds[sign] = n
	return n
}

func newLine(indent int, bld io.Writer) {
	writeByte(bld, '\n')
	for n := 0; n < indent; n++ {
		writeByte(bld, ' ')
	}
}

func namespace(name string) []string {
	parts := strings.Split(name, `::`)
	return parts[:len(parts)-1]
}

func nsName(ns []string, name string) string {
	parts := strings.Split(name, `::`)
	if isParent(ns, parts) {
		return strings.Join(parts[len(ns):], `.`)
	}
	return strings.Join(parts, `.`)
}

func isParent(ns, n []string) bool {
	top := len(ns)
	if top < len(n) {
		for idx := 0; idx < top; idx++ {
			if n[idx] != ns[idx] {
				return false
			}
		}
		return true
	}
	return false
}
