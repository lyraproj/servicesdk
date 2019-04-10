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
	Activity

	// IterationStyle returns the style of iterator, times, range, each, or eachPair.
	IterationStyle() IterationStyle

	// Producer returns the Activity that will be invoked once for each iteration
	Producer() Activity

	// Over returns what this iterator will iterate over.
	Over() px.Value

	// Variables returns the variables that this iterator will produce for each iteration. These
	// variables will be removed from the declared input set when the final requirements
	// for the activity are computed.
	Variables() []px.Parameter
}

type iterator struct {
	activity
	style     IterationStyle
	producer  Activity
	over      px.Value
	variables []px.Parameter
}

func MakeIterator(name string, when Condition, input, output []px.Parameter, style IterationStyle, producer Activity, over px.Value, variables []px.Parameter) Iterator {
	return &iterator{activity{name, when, input, output}, style, producer, over, variables}
}

func (it *iterator) Label() string {
	return `iterator ` + it.name
}

func (it *iterator) IterationStyle() IterationStyle {
	return it.style
}

func (it *iterator) Producer() Activity {
	return it.producer
}

func (it *iterator) Over() px.Value {
	return it.over
}

func (it *iterator) Variables() []px.Parameter {
	return it.variables
}
