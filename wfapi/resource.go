package wfapi

import "github.com/puppetlabs/go-evaluator/eval"

type StateRetriever interface {
	State(input eval.OrderedMap) (state eval.PuppetObject, err error)
}

type Resource interface {
	Activity

	State() StateRetriever
}
