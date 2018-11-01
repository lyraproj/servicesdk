package service

import "github.com/puppetlabs/go-evaluator/eval"

type Metadata interface {
	Metadata() (typeSet eval.TypeSet, definitions []Definition)
}
