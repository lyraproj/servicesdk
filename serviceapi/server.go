package serviceapi

import "github.com/puppetlabs/go-evaluator/eval"

type Service interface {
	Invokable
	Metadata
	StateResolver

	Identifier() eval.TypedName
}
