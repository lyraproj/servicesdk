package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

type IterationStyle int

const IterationStyleEach = 1
const IterationStyleEachPair = 2
const IterationStyleRange = 3
const IterationStyleTimes = 4

func (is IterationStyle) String() string {
	switch is {
	case IterationStyleEach:
		return `each`
	case IterationStyleEachPair:
		return `eachPair`
	case IterationStyleRange:
		return `range`
	case IterationStyleTimes:
		return `times`
	default:
		return `unknown iteration style`
	}
}

func NewIterationStyle(style string) IterationStyle {
	switch style {
	case `each`:
		return IterationStyleEach
	case `eachPair`:
		return IterationStyleEachPair
	case `range`:
		return IterationStyleRange
	case `times`:
		return IterationStyleTimes
	}
	panic(px.Error(IllegalIterationStyle, issue.H{`style`: style}))
}

type Iterator interface {
	Step

	// IterationStyle returns the style of iterator, times, range, each, or eachPair.
	IterationStyle() IterationStyle

	// Producer returns the Step that will be invoked once for each iteration
	Producer() Step

	// Over returns what this iterator will iterate over.
	Over() px.Value

	// Variables returns the variables that this iterator will produce for each iteration. These
	// variables will be removed from the declared parameters set when the final requirements
	// for the step are computed.
	Variables() []px.Parameter

	// Into names the returns from the iteration
	Into() string
}

type iterator struct {
	step
	style     IterationStyle
	producer  Step
	over      px.Value
	variables []px.Parameter
	into      string
}

func MakeIterator(name string, origin issue.Location, when Condition, parameters, returns []px.Parameter,
	style IterationStyle, producer Step, over px.Value, variables []px.Parameter, into string) Iterator {
	return &iterator{step{name, origin, when, parameters, returns}, style, producer, over, variables, into}
}

func (it *iterator) Label() string {
	return `iterator ` + it.name
}

func (it *iterator) IterationStyle() IterationStyle {
	return it.style
}

func (it *iterator) Producer() Step {
	return it.producer
}

func (it *iterator) Over() px.Value {
	return it.over
}

func (it *iterator) Into() string {
	return it.into
}

func (it *iterator) Variables() []px.Parameter {
	return it.variables
}
