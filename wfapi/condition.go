package wfapi

import (
	"fmt"
	"github.com/lyraproj/pcore/px"
)

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
var Boolean func(bool) Condition

// Truthy returns a Condition that yields true when the variable
// named by the given name contains a truthy value (i.e. not undef or false)
var Truthy func(string) Condition

// Not returns a Condition that yields true when the given condition
// yields false
var Not func(Condition) Condition

// And returns a Condition that yields true when all given conditions
// yield true
var Or func([]Condition) Condition

// Or returns a Condition that yields true when at least one of the given conditions
// yields true
var And func([]Condition) Condition
