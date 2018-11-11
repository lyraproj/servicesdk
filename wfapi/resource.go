package wfapi

import "github.com/puppetlabs/go-evaluator/eval"

type Resource interface {
	Activity

	State() eval.PuppetObject
}
