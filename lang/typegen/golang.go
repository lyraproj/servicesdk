package typegen

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lyraproj/issue/issue"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/wf"
)

type goGeneratorFactory struct {
}

func (gf *goGeneratorFactory) GenerateTypes(typeSet px.TypeSet, directory string) {
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
		pkg := strings.ToLower(wf.LeafName(typeSet.Name()))
		formattedTypeToStream(typeSet, directory, pkg, func(g *goGenerator, b *bytes.Buffer) {
			for _, t := range tts {
				b.WriteByte('\n')
				g.generateType(g.goTypeName(t), t, 0, b)
			}
		})
	}
}

func (g *goGenerator) goTypeName(t px.Type) string {
	n := wf.LeafName(t.Name())
	if g.useCamelCase {
		n = issue.SnakeToCamelCase(n)
	}
	return n
}

func (gf *goGeneratorFactory) GenerateType(typ px.Type, directory string) {
	sg := strings.Split(typ.Name(), `::`)
	pkg := ``
	if len(sg) > 1 {
		pkg = strings.ToLower(sg[len(sg)-2])
	}
	formattedTypeToStream(typ, directory, pkg, func(g *goGenerator, b *bytes.Buffer) {
		b.WriteByte('\n')
		g.generateType(g.goTypeName(typ), typ, 0, b)
	})
}

func formattedTypeToStream(t px.Type, directory string, pkg string, f func(g *goGenerator, b *bytes.Buffer)) {
	tsp := append(strings.Split(strings.ToLower(t.Name()), `::`), pkg)
	directory = filepath.Join(directory, filepath.Join(tsp...)) + `.go`
	typeToStream(directory, func(w io.Writer) {
		g := makeGoGenerator(pkg)
		g.findAnonymousTypes(t, nil)
		g.nameAnonymousTypes()
		b := bytes.NewBufferString("// this file is generated\n")
		b.WriteString(`package `)
		b.WriteString(pkg)
		b.WriteString("\n\nimport (")
		newLine(1, b)
		b.WriteString(`"fmt"`)
		newLine(1, b)
		b.WriteString(`"reflect"`)
		if g.includeRegexp {
			newLine(1, b)
			b.WriteString(`"regexp"`)
		}
		if g.includeTime {
			newLine(1, b)
			b.WriteString(`"time"`)
		}
		newLine(0, b)
		newLine(1, b)
		b.WriteString(`"github.com/lyraproj/pcore/px"`)
		if g.includeSemver {
			newLine(1, b)
			b.WriteString(`"github.com/lyraproj/semver/semver"`)
		}
		newLine(0, b)
		b.WriteByte(')')
		f(g, b)
		b.WriteByte('\n')
		g.writeAnonymousTypes(b)
		g.writeInit(b)
		_, err := w.Write(g.formatCode(b.Bytes()))
		if err != nil {
			panic(err)
		}
	})
}

type goGenerator struct {
	anonTypes     []*anonType
	allTypes      map[string]px.Type
	anonNames     map[string]bool
	pkg           string
	useCamelCase  bool
	includeTime   bool
	includeRegexp bool
	includeSemver bool
}

func makeGoGenerator(pkg string) *goGenerator {
	return &goGenerator{
		pkg:          pkg,
		anonTypes:    make([]*anonType, 0, 50),
		allTypes:     make(map[string]px.Type, 100),
		anonNames:    make(map[string]bool, 100),
		useCamelCase: true,
	}
}

type nameSeg struct {
	n string // The segment
	w int    // How many paths that uses this segment at the same position
}

func (n *nameSeg) String() string {
	return fmt.Sprintf(`%s(%d)`, n.n, n.w)
}

type anonType struct {
	t  px.Type    // The anonymous type (a Struct or an Object
	ps [][]string // All paths that lead up to this type, in reverse
	cc bool       // true when generator resorted to concatenate two first names in each path
	n  string     // Generated name
}

func (a *anonType) String() string {
	return fmt.Sprintf(`%v, %s`, a.ps, a.n)
}

func (a *anonType) mostCommonPath() []*nameSeg {
	mc := a.findCommonPath(false)
	if len(mc) == 0 && !a.cc {
		mc = a.findCommonPath(true)
	}
	return mc
}

// findCommonPath attempts to find a path common to all paths that leads to this
// anonType. If no such path is found, the most common segment at each position
// is used to form the path. If concatAlt is true, then if no most common segment
// is fount, then segments that are found  to have equal number of occurrences for a
// specific path position can be concatenated to resolve the conflict.
func (a *anonType) findCommonPath(concatAlt bool) (result []*nameSeg) {
	for s := 0; ; s++ {
		maxCount := 0
		maxName := ``
		counts := make(map[string]int, len(a.ps))
		for _, pe := range a.ps {
			if s >= len(pe) {
				return result
			}
			n := pe[s]
			v, ok := counts[n]
			if ok {
				v++
			} else {
				v = 1
			}
			counts[n] = v
			if v > maxCount {
				maxCount = v
				maxName = n
			}
		}

		if len(counts) > 1 {
			// Check that maxCount is unique
			if concatAlt {
				sb := bytes.NewBufferString(``)
				for n, mx := range counts {
					if mx == maxCount {
						sb.WriteString(n)
					}
				}
				maxName = sb.String()
			} else {
				mxFound := false
				for _, mx := range counts {
					if mx == maxCount {
						if mxFound {
							return result
						}
						mxFound = true
					}
				}
			}
		}
		result = append(result, &nameSeg{maxName, maxCount})
	}
}

func (a *anonType) desiredName(allNames map[string]px.Type) (string, int) {
	mc := a.mostCommonPath()
	b := bytes.NewBufferString(``)
	for _, ns := range mc {
		s := b.String()
		b.Reset()
		b.WriteString(ns.n)
		b.WriteString(s)
		s = b.String()
		if len(s) > 2 {
			if _, ok := allNames[s]; !ok {
				return s, ns.w
			}
		}
	}
	return ``, 0
}

// useFirstUniqueName sorts all paths to the anonymous type, shortest path
// first, and then makes an attempt to form a unique name from each path, starting
// with segment 0, then concatenating with segment 1, then 2, etc. The first name
// that is formed that doesn't already exist in the allNames map, is assigned as
// the name of the entry and added to the allNames map.
//
// This method should only be used as the last resort.
func (a *anonType) useFirstUniqueName(allNames map[string]px.Type) bool {
	b := bytes.NewBufferString(``)
	ps := a.ps
	sort.Slice(ps, func(i, j int) bool {
		return len(ps[i]) < len(ps[j])
	})

	for _, p := range ps {
		b.Reset()
		for _, n := range p {
			s := b.String()
			b.Reset()
			b.WriteString(n)
			b.WriteString(s)
			s = b.String()
			if len(s) > 2 {
				if _, ok := allNames[s]; !ok {
					allNames[s] = a.t
					a.n = s
					return true
				}
			}
		}
	}
	return false
}

func (g *goGenerator) appendTypeWithPath(t px.Type, p []string) {
	l := len(p)
	rp := make([]string, 0, l)
	for i := l - 1; i >= 0; i-- {
		vs := strings.Split(p[i], `_`)
		for vi := len(vs) - 1; vi >= 0; vi-- {
			rp = append(rp, strings.Title(vs[vi]))
		}
	}
	for _, ot := range g.anonTypes {
		if t.Equals(ot.t, nil) {
			ot.ps = append(ot.ps, rp)
			return
		}
	}
	g.anonTypes = append(g.anonTypes, &anonType{t, [][]string{rp}, false, ``})
}

func (g *goGenerator) anonymousName(t px.Type) string {
	for _, ot := range g.anonTypes {
		if t.Equals(ot.t, nil) {
			return ot.n
		}
	}
	panic(fmt.Errorf(`unable to find generated name for anonymous type %s`, px.ToPrettyString(t)))
}

func (g *goGenerator) nameAnonymousTypes() {
	stuffHappened := true
	as := g.anonTypes
	for stuffHappened {
		stuffHappened = false

		// Build map of desired names and what weigh that each anonymous type
		// have for the desired name.
		nwsMap := make(map[string]map[int][]*anonType, len(as))
		keys := make([]string, 0, len(as))
		for _, a := range as {
			if a.n != `` {
				continue
			}

			cn, weight := a.desiredName(g.allTypes)
			if weight == 0 {
				// No name produced. We'll deal with it later
				continue
			}
			if nws, ok := nwsMap[cn]; ok {
				if ns, ok := nws[weight]; ok {
					nws[weight] = append(ns, a)
				} else {
					nws[weight] = []*anonType{a}
				}
			} else {
				nwsMap[cn] = map[int][]*anonType{weight: {a}}
				keys = append(keys, cn)
			}
		}

		// Assign names to all entries that only have one weight
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})
		for _, n := range keys {
			nws := nwsMap[n]
			var a *anonType
			if len(nws) == 1 {
				// Must use range to pick the one and only element
				for _, ns := range nws {
					// First entry wins the name in case of a weight conflict
					a = ns[0]
					break
				}
			} else {
				// Find type with heaviest weight
				maxWeight := 0
				for w, ns := range nws {
					if w > maxWeight {
						// First entry wins the name in case of a weight conflict
						a = ns[0]
						maxWeight = w
					}
				}
			}
			if a != nil {
				a.n = n
				g.allTypes[n] = a.t
				stuffHappened = true
			}
		}

		if stuffHappened {
			// Above code must be reiterated
			continue
		}

		for _, a := range as {
			// Modify all entries by concatenating the two first entries in each
			// path of the anonType. This is a one time operation on each entry
			if a.n == `` && !a.cc {
				allConcat := true
				for _, p := range a.ps {
					if len(p) < 2 {
						allConcat = false
						break
					}
				}
				if allConcat {
					for i, p := range a.ps {
						a.ps[i] = append([]string{p[1] + p[0]}, p[2:]...)
					}
					a.cc = true
					stuffHappened = true
				}
			}
		}

		if stuffHappened {
			// Above code must be reiterated
			continue
		}

		// As a last resort, try making a unique name from one of the paths.
		for _, a := range as {
			if a.n == `` {
				if a.useFirstUniqueName(g.allTypes) {
					stuffHappened = true
				}
			}
		}
	}

	sort.Slice(as, func(i, j int) bool {
		a := as[i]
		b := as[j]
		if a.n != `` {
			if b.n != `` {
				return a.n < b.n
			}
			return true
		}
		if b.n != `` {
			return false
		}
		return strings.Join(a.ps[0], `/`) < strings.Join(b.ps[0], `/`)
	})

	for _, a := range as {
		if a.n == `` {
			panic(fmt.Errorf(`unable to generate name for %s`, a))
		}
		g.anonNames[a.n] = true
	}
}

func (g *goGenerator) findAnonymousTypes(t px.Type, p []string) {
	switch t := t.(type) {
	case px.TypeSet:
		t.Types().EachPair(func(k, v px.Value) {
			g.findAnonymousTypes(v.(px.Type), append(p, k.String()))
		})
	case px.ObjectType:
		n := t.Name()
		if n != `` {
			n := g.goTypeName(t)
			if _, ok := g.allTypes[n]; ok {
				break
			}
			g.allTypes[n] = t
		}
		for _, a := range t.AttributesInfo().Attributes() {
			g.findAnonymousTypes(a.Type(), append(p, a.Name()))
		}
		if t.Name() == `` {
			g.appendTypeWithPath(t, p)
		}
	case px.TypeWithContainedType:
		g.findAnonymousTypes(t.ContainedType(), p)
	case *types.ArrayType:
		g.findAnonymousTypes(t.ElementType(), p)
	case *types.HashType:
		g.findAnonymousTypes(t.KeyType(), p)
		g.findAnonymousTypes(t.ValueType(), p)
	case *types.StructType:
		for _, se := range t.Elements() {
			g.findAnonymousTypes(se.Key(), p)
			g.findAnonymousTypes(se.Value(), append(p, se.Name()))
			g.appendTypeWithPath(t, p)
		}
	case *types.TupleType:
		for _, vt := range t.Types() {
			g.findAnonymousTypes(vt, p)
		}
	case *types.VariantType:
		for _, vt := range t.Types() {
			g.findAnonymousTypes(vt, p)
		}
	case *types.RegexpType:
		g.includeRegexp = true
	case *types.SemVerType:
		g.includeSemver = true
	case *types.SemVerRangeType:
		g.includeSemver = true
	case *types.TimestampType:
		g.includeTime = true
	case *types.TimespanType:
		g.includeTime = true
	}
	// TODO: Include packages of Object types from other TypeSets
}

func (g *goGenerator) generateType(name string, t px.Type, i int, b *bytes.Buffer) {
	switch t := t.(type) {
	case px.ObjectType:
		g.generateObjectType(name, t, 0, b)
	case *types.StructType:
		g.generateStructType(name, t, 0, b)
	default:
		b.WriteString(`type `)
		b.WriteString(name)
		b.WriteString(` = `)
		g.writeType(t, b)
	}
}

func (g *goGenerator) generateObjectType(name string, t px.ObjectType, indent int, b *bytes.Buffer) {
	newLine(indent, b)
	b.WriteString("type ")
	b.WriteString(name)
	b.WriteString(` struct {`)
	indent++
	for _, a := range t.AttributesInfo().Attributes() {
		newLine(indent, b)
		n := a.Name()
		if g.useCamelCase {
			n = issue.SnakeToCamelCase(n)
		} else {
			n = strings.Title(n)
		}
		b.WriteString(n)
		b.WriteByte(' ')
		if a.HasValue() {
			b.WriteByte('*')
		}
		g.writeRequiredType(a.Type(), b)
		if g.useCamelCase && issue.FirstToLower(n) != a.Name() {
			b.WriteString(" `puppet:\"name=>'")
			b.WriteString(a.Name())
			b.WriteString("'\"`")
		}
	}
	indent--
	newLine(indent, b)
	b.WriteString(`}`)
}

func (g *goGenerator) generateStructType(name string, t *types.StructType, indent int, b *bytes.Buffer) {
	b.WriteString("\ntype ")
	b.WriteString(name)
	b.WriteString(` struct {`)
	indent++
	for _, e := range t.Elements() {
		newLine(indent, b)
		b.WriteString(strings.Title(e.Name()))
		b.WriteString(` `)
		g.writeType(e.Value(), b)
	}
	indent--
	newLine(indent, b)
	b.WriteString(`}`)
}

func (g *goGenerator) writeAnonymousTypes(b *bytes.Buffer) {
	for _, a := range g.anonTypes {
		b.WriteByte('\n')
		g.generateType(a.n, a.t, 0, b)
	}
}

func (g *goGenerator) writeInit(b *bytes.Buffer) {
	indent := 0
	b.WriteString("\nfunc InitTypes(c px.Context) {")
	indent++
	rts := make([]px.ObjectType, 0, len(g.allTypes)-len(g.anonNames))
	for k, t := range g.allTypes {
		if _, ok := g.anonNames[k]; ok {
			continue
		}
		if rt, ok := t.(px.ObjectType); ok {
			rts = append(rts, rt)
		}
	}
	sort.Slice(rts, func(i, j int) bool {
		return rts[i].Name() < rts[j].Name()
	})

	newLine(indent, b)
	b.WriteString(`load := func(n string) px.Type {`)
	indent++
	newLine(indent, b)
	b.WriteString(`if v, ok := px.Load(c, px.NewTypedName(px.NsType, n)); ok {`)
	newLine(indent+1, b)
	b.WriteString(`return v.(px.Type)`)
	newLine(indent, b)
	b.WriteByte('}')
	newLine(indent, b)
	b.WriteString(`panic(fmt.Errorf("unable to load Type '%s'", n))`)
	indent--
	newLine(indent, b)
	b.WriteString("}\n")
	newLine(indent, b)
	b.WriteString(`ir := c.ImplementationRegistry()`)
	for _, rt := range rts {
		newLine(indent, b)
		b.WriteString(`ir.RegisterType(load("`)
		b.WriteString(rt.Name())
		b.WriteString(`"), reflect.TypeOf(&`)
		b.WriteString(g.goTypeName(rt))
		b.WriteString(`{}))`)
	}

	indent--
	newLine(indent, b)
	b.WriteString(`}`)
}

func (g *goGenerator) writeType(t px.Type, b *bytes.Buffer) {
	if ot, ok := t.(*types.OptionalType); ok {
		b.WriteByte('*')
		t = ot.ContainedType()
	}
	g.writeRequiredType(t, b)
}

func (g *goGenerator) writeRequiredType(t px.Type, b *bytes.Buffer) {
	switch t := t.(type) {
	case *types.OptionalType:
		g.writeRequiredType(t.ContainedType(), b)
	case *types.ArrayType:
		b.WriteString("[]")
		g.writeType(t.ElementType(), b)
	case *types.TupleType:
		b.WriteString("[]")
		g.writeType(t.CommonElementType(), b)
	case *types.HashType:
		b.WriteString("map[")
		g.writeType(t.KeyType(), b)
		b.WriteByte(']')
		g.writeType(t.ValueType(), b)
	case *types.StructType:
		b.WriteString(g.anonymousName(t))
	case px.ObjectType:
		n := t.Name()
		if n == `` {
			n = g.anonymousName(t)
		} else {
			n = g.goTypeName(t)
		}
		b.WriteString(n)
	case *types.BooleanType:
		b.WriteString("bool")
	case *types.IntegerType:
		b.WriteString("int64")
	case *types.FloatType:
		b.WriteString("float64")
	case px.StringType, *types.EnumType, *types.PatternType:
		b.WriteString("string")
	case *types.RegexpType:
		b.WriteString("regexp.Regexp")
	case *types.SemVerType:
		b.WriteString("semver.Version")
	case *types.SemVerRangeType:
		b.WriteString("semver.VersionRange")
	case *types.TimestampType:
		b.WriteString("time.Time")
	case *types.TimespanType:
		b.WriteString("time.Duration")
	default:
		panic(fmt.Errorf("don't know how to generate Go type from: %s", px.ToPrettyString(t)))
	}
}

// formatCode reformats the code as `go fmt` would
func (g *goGenerator) formatCode(code []byte) []byte {
	src, err := format.Source(code)
	if err != nil {
		panic(fmt.Errorf("unexpected error running format.Source: %s", err.Error()))
	}
	return src
}
