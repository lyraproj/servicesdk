package wfapi

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-issues/issue"
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
	case `detete`:
		return Delete
	case `upsert`:
		return Upsert
	}
	panic(eval.Error(WF_ILLEGAL_OPERATION, issue.H{`operation`: operation}))
}
