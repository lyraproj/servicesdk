package wfapi

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
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
	panic(eval.Error(WF_ILLEGAL_ITERATION_STYLE, issue.H{`style`: style}))
}

type Iterator interface {
	Activity

	// Style returns the style of iterator, times, range, each, or eachPair.
	IterationStyle() IterationStyle

	// Producer returns the Activity that will be invoked once for each iteration
	Producer() Activity

	// Over returns what this iterator will iterate over. These parameters will be added
	// to the declared input set when the final requirements for the activity are computed.
	Over() []eval.Parameter

	// Variables returns the variables that this iterator will produce for each iteration. These
	// variables will be removed from the declared input set when the final requirements
	// for the activity are computed.
	Variables() []eval.Parameter
}
