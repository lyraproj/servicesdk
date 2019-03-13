package wf

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/lyraproj/pcore/px"
)

type boolean bool

const Always = boolean(true)
const Never = boolean(false)

// A Condition evaluates to true or false depending on its given input
type Condition interface {
	fmt.Stringer

	// Precedence returns the operator precedence for this Condition
	Precedence() int

	// IsTrue returns true if the given input satisfies the condition, false otherwise
	IsTrue(input px.OrderedMap) bool

	// Returns all names in use by this condition and its nested conditions. The returned
	// slice is guaranteed to be unique and sorted alphabetically
	Names() []string
}

// Boolean returns that Condition that yields the given boolean
func Boolean(v bool) Condition {
	return boolean(v)
}

func (b boolean) String() string {
	if b {
		return `true`
	}
	return `false`
}

func (b boolean) Precedence() int {
	return 5
}

func (b boolean) IsTrue(input px.OrderedMap) bool {
	return bool(b)
}

func (b boolean) Names() []string {
	return []string{}
}

type truthy string

// Truthy returns a Condition that yields true when the variable
// named by the given name contains a truthy value (i.e. not undef or false)
func Truthy(name string) Condition {
	return truthy(name)
}

func (v truthy) IsTrue(input px.OrderedMap) bool {
	value, ok := input.Get4(string(v))
	return ok && px.IsTruthy(value)
}

func (v truthy) Names() []string {
	return []string{string(v)}
}

func (v truthy) Precedence() int {
	return 4
}

func (v truthy) String() string {
	return string(v)
}

// Not returns a Condition that yields true when the given condition
// yields false
func Not(condition Condition) Condition {
	return &not{condition}
}

type not struct {
	condition Condition
}

func (n *not) IsTrue(input px.OrderedMap) bool {
	return !n.condition.IsTrue(input)
}

func (n *not) Names() []string {
	return n.condition.Names()
}

func (n *not) Precedence() int {
	return 3
}

func (n *not) String() string {
	b := bytes.NewBufferString(`!`)
	emitContained(n.condition, n.Precedence(), b)
	return b.String()
}

type and struct {
	conditions []Condition
}

// And returns a Condition that yields true when all given conditions
// yield true
func And(conditions []Condition) Condition {
	return &and{conditions}
}

func (a *and) IsTrue(input px.OrderedMap) bool {
	for _, condition := range a.conditions {
		if !condition.IsTrue(input) {
			return false
		}
	}
	return true
}

func (a *and) Names() []string {
	return mergeNames(a.conditions)
}

func (a *and) Precedence() int {
	return 2
}

func (a *and) String() string {
	return concat(a.conditions, a.Precedence(), `and`)
}

// Or returns a Condition that yields true when at least one of the given conditions
// yield true
func Or(conditions []Condition) Condition {
	return &or{conditions}
}

type or struct {
	conditions []Condition
}

func (o *or) IsTrue(input px.OrderedMap) bool {
	for _, condition := range o.conditions {
		if condition.IsTrue(input) {
			return true
		}
	}
	return false
}

func (o *or) Names() []string {
	return mergeNames(o.conditions)
}

func (o *or) Precedence() int {
	return 1
}

func (o *or) String() string {
	return concat(o.conditions, o.Precedence(), `or`)
}

func mergeNames(conditions []Condition) []string {
	h := make(map[string]bool)
	for _, c := range conditions {
		for _, n := range c.Names() {
			h[n] = true
		}
	}
	names := make([]string, 0, len(h))
	for n := range h {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func concat(conditions []Condition, precedence int, op string) string {
	b := bytes.NewBufferString(``)
	for i, c := range conditions {
		if i > 0 {
			b.WriteByte(' ')
			b.WriteString(op)
			b.WriteByte(' ')
		}
		emitContained(c, precedence, b)
	}
	return b.String()
}

func emitContained(c Condition, p int, b *bytes.Buffer) {
	if p > c.Precedence() {
		b.WriteByte('(')
		b.WriteString(c.String())
		b.WriteByte(')')
	} else {
		b.WriteString(c.String())
	}
}
