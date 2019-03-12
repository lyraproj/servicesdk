package wf

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

type Operation int

const Read Operation = 1
const Delete Operation = 2
const Upsert Operation = 3

func (is Operation) String() string {
	switch is {
	case Read:
		return `read`
	case Delete:
		return `delete`
	case Upsert:
		return `upsert`
	default:
		return `unknown operation`
	}
}

func NewOperation(operation string) Operation {
	switch operation {
	case `read`:
		return Read
	case `delete`:
		return Delete
	case `upsert`:
		return Upsert
	}
	panic(px.Error(IllegalOperation, issue.H{`operation`: operation}))
}
