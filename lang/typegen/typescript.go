package typegen

import (
	"bytes"
	"fmt"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/puppet-evaluator/utils"
	"io"
	"strings"
)

type tsGenerator struct{}

func (g *tsGenerator) GenerateTypes(typeSet eval.TypeSet, directory string) {
	hasNonTypeSetTypes := false
	typeSet.Types().EachValue(func(t eval.Value) {
		if ts, ok := t.(eval.TypeSet); ok {
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

func (g *tsGenerator) GenerateType(typ eval.Type, directory string) {
	typeToStream(typ, directory, `.ts`, func(b io.Writer) {
		write(b, "// this file is generated\n")
		write(b, "import {PcoreValue, Value} from 'lyra-workflow';")
		g.generateType(typ, namespace(typ.Name()), 0, b)
	})
}

// GenerateTypes produces TypeScript types for all types in the given TypeSet and appends them to
// the given buffer.
func (g *tsGenerator) generateTypes(ts eval.TypeSet, ns []string, indent int, bld io.Writer) {
	newLine(indent, bld)
	leafName := nsName(ns, ts.Name())
	ns = append(ns, leafName)
	ts.Types().EachValue(func(t eval.Value) { g.generateType(t.(eval.Type), ns, indent, bld) })
}

// GenerateType produces a TypeScript type for the given Type and appends it to
// the given buffer.
func (g *tsGenerator) generateType(t eval.Type, ns []string, indent int, bld io.Writer) {
	if _, ok := t.(eval.TypeSet); ok {
		return
	}

	if pt, ok := t.(eval.ObjectType); ok {
		newLine(indent, bld)
		write(bld, `export class `)
		write(bld, nsName(ns, pt.Name()))
		if ppt, ok := pt.Parent().(eval.ObjectType); ok {
			write(bld, ` extends `)
			write(bld, nsName(ns, ppt.Name()))
		} else {
			write(bld, ` implements PcoreValue`)
		}
		write(bld, ` {`)
		indent += 2
		ai := pt.AttributesInfo()
		allAttrs, thisAttrs, superAttrs := g.toTsAttrs(pt, ns, ai.Attributes())
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
func (g *tsGenerator) ToTsType(ns []string, pType eval.Type) string {
	bld := bytes.NewBufferString(``)
	appendTsType(ns, pType, bld)
	return bld.String()
}

type tsAttribute struct {
	name  string
	typ   string
	value *string
}

func (g *tsGenerator) toTsAttrs(t eval.ObjectType, ns []string, attrs []eval.Attribute) (allAttrs, thisAttrs, superAttrs []*tsAttribute) {
	allAttrs = make([]*tsAttribute, len(attrs))
	superAttrs = make([]*tsAttribute, 0)
	thisAttrs = make([]*tsAttribute, 0)
	for i, attr := range attrs {
		tsAttr := &tsAttribute{name: attr.Name(), typ: g.ToTsType(ns, attr.Type())}
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
		write(bld, attr.name)
		write(bld, `: `)
		write(bld, attr.typ)
		write(bld, `;`)
	}
	return
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
			write(bld, attr.name)
			write(bld, `: `)
			write(bld, attr.name)
		}
		write(bld, `});`)
	}
	for _, attr := range thisAttrs {
		newLine(indent, bld)
		write(bld, `this.`)
		write(bld, attr.name)
		write(bld, ` = `)
		write(bld, attr.name)
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
				write(bld, attr.name)
				write(bld, ` !== `)
				write(bld, *attr.value)
				write(bld, `) {`)
				indent += 2
				newLine(indent, bld)
			}
			write(bld, `ih['`)
			write(bld, attr.name)
			write(bld, `'] = this.`)
			write(bld, attr.name)
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
		write(bld, attr.name)
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
		write(bld, attr.name)
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

func toTsValue(value eval.Value) *string {
	bld := bytes.NewBufferString(``)
	appendTsValue(value, bld)
	s := bld.String()
	return &s
}

func appendTsValue(value eval.Value, bld io.Writer) {
	switch value.(type) {
	case *types.UndefValue:
		write(bld, `null`)
	case eval.StringValue:
		utils.PuppetQuote(bld, value.String())
	case eval.BooleanValue, eval.IntegerValue, eval.FloatValue:
		write(bld, value.String())
	case *types.ArrayValue:
		writeByte(bld, '[')
		value.(*types.ArrayValue).EachWithIndex(func(e eval.Value, i int) {
			if i > 0 {
				write(bld, `, `)
			}
			appendTsValue(e, bld)
		})
		writeByte(bld, ']')
	case *types.HashValue:
		writeByte(bld, '{')
		value.(*types.HashValue).EachWithIndex(func(e eval.Value, i int) {
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

func appendTsType(ns []string, pType eval.Type, bld io.Writer) {
	switch pType.(type) {
	case *types.BooleanType:
		write(bld, `boolean`)
	case *types.IntegerType, *types.FloatType:
		write(bld, `number`)
	case eval.StringType:
		write(bld, `string`)
	case *types.OptionalType:
		appendTsType(ns, pType.(*types.OptionalType).ContainedType(), bld)
		write(bld, `|null`)
	case *types.ArrayType:
		appendTsType(ns, pType.(*types.ArrayType).ElementType(), bld)
		write(bld, `[]`)
	case *types.VariantType:
		for i, v := range pType.(*types.VariantType).Types() {
			if i > 0 {
				write(bld, `|`)
			}
			appendTsType(ns, v, bld)
		}
	case *types.HashType:
		ht := pType.(*types.HashType)
		write(bld, `{[s: `)
		appendTsType(ns, ht.KeyType(), bld)
		write(bld, `]: `)
		appendTsType(ns, ht.ValueType(), bld)
		write(bld, `}`)
	case *types.EnumType:
		for i, s := range pType.(*types.EnumType).Parameters() {
			if i > 0 {
				write(bld, `|`)
			}
			appendTsValue(s, bld)
		}
	case *types.TypeAliasType:
		write(bld, nsName(ns, pType.(*types.TypeAliasType).Name()))
	case eval.ObjectType:
		write(bld, nsName(ns, pType.(eval.ObjectType).Name()))
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

func relativeNs(ns []string, name string) []string {
	parts := strings.Split(name, `::`)
	if len(parts) == 1 {
		return []string{}
	}
	if len(ns) == 0 || isParent(ns, parts) {
		return parts[len(ns) : len(parts)-1]
	}
	panic(fmt.Errorf("cannot generate %s in namespace %s", name, ns))
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
