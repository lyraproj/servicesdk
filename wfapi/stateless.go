package wfapi

import "github.com/puppetlabs/go-evaluator/eval"

type Doer interface {
	Do(op Operation, input eval.OrderedMap) (output eval.OrderedMap, err error)
}

type Stateless interface {
	Activity

	Interface() Doer
}
