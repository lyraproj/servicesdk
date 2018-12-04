package serviceapi

import "github.com/lyraproj/puppet-evaluator/eval"

type Service interface {
	Invokable
	Metadata
	StateResolver

	Identifier(eval.Context) eval.TypedName
}
