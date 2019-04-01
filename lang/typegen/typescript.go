package typegen

import (
	"bytes"
	"io"
	"strings"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/pcore/utils"
)

type tsGenerator struct{}

func (g *tsGenerator) GenerateTypes(typeSet px.TypeSet, directory string) {
	hasNonTypeSetTypes := false
	typeSet.Types().EachValue(func(t px.Value) {
		if ts, ok := t.(px.TypeSet); ok {
			g.GenerateTypes(ts, directory)
		} else {
			hasNonTypeSetTypes = true
		}
	})

	if hasNonTypeSetTypes {
		typeToStream(typeSet, directory, `.ts`, func(b io.Writer) {
			write(b, "// this file is generated\n")
			write(b, "import {PcoreValue, Value} from 'lyra-workflow';")
			g.generateTypes(typeSet, namespace(typeSet.Name()), 0, b)
		})
	}
}

func (g *tsGenerator) GenerateType(typ px.Type, directory string) {
	typeToStream(typ, directory, `.ts`, func(b io.Writer) {
		write(b, "// this file is generated\n")
		write(b, "import {PcoreValue, Value} from 'lyra-workflow';")
		g.generateType(typ, namespace(typ.Name()), 0, b)
	})
}

// GenerateTypes produces TypeScript types for all types in the given TypeSet and appends them to
// the given buffer.
func (g *tsGenerator) generateTypes(ts px.TypeSet, ns []string, indent int, bld io.Writer) {
	newLine(indent, bld)
	leafName := nsName(ns, ts.Name())
	ns = append(ns, leafName)
	ts.Types().EachValue(func(t px.Value) { g.generateType(t.(px.Type), ns, indent, bld) })
}

// GenerateType produces a TypeScript type for the given Type and appends it to
// the given buffer.
func (g *tsGenerator) generateType(t px.Type, ns []string, indent int, bld io.Writer) {
	if _, ok := t.(px.TypeSet); ok {
		return
	}

	if pt, ok := t.(px.ObjectType); ok {
		newLine(indent, bld)
		write(bld, `export class `)
		write(bld, nsName(ns, pt.Name()))
		if ppt, ok := pt.Parent().(px.ObjectType); ok {
			write(bld, ` extends `)
			write(bld, nsName(ns, ppt.Name()))
		} else {
			write(bld, ` implements PcoreValue`)
		}
		write(bld, ` {`)
		indent += 2
		ai := pt.AttributesInfo()
		allAttrs, thisAttrs, superAttrs := toTsAttrs(pt, ns, ai.Attributes())
		appendFields(thisAttrs, indent, bld)
		if len(thisAttrs) > 0 {
			writeByte(bld, '\n')
		}
		if len(allAttrs) > 0 {
			appendConstructor(allAttrs, thisAttrs, superAttrs, indent, bld)
			writeByte(bld, '\n')
		}
		hasSuper := len(superAttrs) > 0
		if len(thisAttrs) > 0 || !hasSuper {
			appendPValueGetter(hasSuper, thisAttrs, indent, bld)
			writeByte(bld, '\n')
		}
		appendPTypeGetter(pt.Name(), indent, bld)
		indent -= 2
		newLine(indent, bld)
		write(bld, "}\n")
	} else {
		appendTsType(ns, t, bld)
	}
}

// ToTsType converts the given pType to a string representation of a TypeScript type. The given
// pType can not be a TypeSet.
func (g *tsGenerator) ToTsType(ns []string, pType px.Type) string {
	return toTsType(ns, pType)
}

func toTsType(ns []string, pType px.Type) string {
	bld := bytes.NewBufferString(``)
	appendTsType(ns, pType, bld)
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

func toTsAttrs(t px.ObjectType, ns []string, attrs []px.Attribute) (allAttrs, thisAttrs, superAttrs []*tsAttribute) {
	allAttrs = make([]*tsAttribute, len(attrs))
	superAttrs = make([]*tsAttribute, 0)
	thisAttrs = make([]*tsAttribute, 0)
	for i, attr := range attrs {
		n := attr.Name()
		tsn := n
		if keywords[n] {
			tsn = n + `_`
		}
		tsAttr := &tsAttribute{tsName: tsn, name: n, typ: toTsType(ns, attr.Type())}
		if attr.HasValue() {
			tsAttr.value = toTsValue(attr.Value())
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

func appendFields(thisAttrs []*tsAttribute, indent int, bld io.Writer) {
	for _, attr := range thisAttrs {
		newLine(indent, bld)
		write(bld, `readonly `)
		write(bld, attr.tsName)
		write(bld, `: `)
		write(bld, attr.typ)
		write(bld, `;`)
	}
}

func appendConstructor(allAttrs, thisAttrs, superAttrs []*tsAttribute, indent int, bld io.Writer) {
	newLine(indent, bld)
	write(bld, `constructor(`)
	appendParameters(allAttrs, indent, bld)
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

func appendPValueGetter(hasSuper bool, thisAttrs []*tsAttribute, indent int, bld io.Writer) {
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

func appendPTypeGetter(name string, indent int, bld io.Writer) {
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

func appendParameters(params []*tsAttribute, indent int, bld io.Writer) {
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

func toTsValue(value px.Value) *string {
	bld := bytes.NewBufferString(``)
	appendTsValue(value, bld)
	s := bld.String()
	return &s
}

func appendTsValue(value px.Value, bld io.Writer) {
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
			appendTsValue(e, bld)
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
			appendTsValue(ev.Value(), bld)
		})
		writeByte(bld, '}')
	}
}

func appendTsType(ns []string, pType px.Type, bld io.Writer) {
	switch pType := pType.(type) {
	case *types.BooleanType:
		write(bld, `boolean`)
	case *types.IntegerType, *types.FloatType:
		write(bld, `number`)
	case px.StringType:
		write(bld, `string`)
	case *types.OptionalType:
		appendTsType(ns, pType.ContainedType(), bld)
		write(bld, `|null`)
	case *types.ArrayType:
		et := pType.ElementType()
		switch et.(type) {
		case *types.ArrayType, *types.EnumType, *types.HashType, *types.OptionalType, *types.VariantType:
			write(bld, `Array<`)
			appendTsType(ns, et, bld)
			write(bld, `>`)
		default:
			appendTsType(ns, et, bld)
			write(bld, `[]`)
		}
	case *types.VariantType:
		for i, v := range pType.Types() {
			if i > 0 {
				write(bld, `|`)
			}
			appendTsType(ns, v, bld)
		}
	case *types.HashType:
		write(bld, `{[s: `)
		appendTsType(ns, pType.KeyType(), bld)
		write(bld, `]: `)
		appendTsType(ns, pType.ValueType(), bld)
		write(bld, `}`)
	case *types.EnumType:
		for i, s := range pType.Parameters() {
			if i > 0 {
				write(bld, `|`)
			}
			appendTsValue(s, bld)
		}
	case *types.TypeAliasType:
		write(bld, nsName(ns, pType.Name()))
	case px.ObjectType:
		if pType.Name() == `` {
			write(bld, `{`)
			allAttrs, _, _ := toTsAttrs(pType, ns, pType.AttributesInfo().Attributes())
			for i, a := range allAttrs {
				if i > 0 {
					write(bld, `,`)
				}
				write(bld, a.tsName)
				write(bld, `: `)
				write(bld, a.typ)
			}
			write(bld, `}`)
		} else {
			write(bld, nsName(ns, pType.Name()))
		}
	}
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
