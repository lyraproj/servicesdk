package serviceapi

import (
	"github.com/puppetlabs/go-evaluator/eval"
)

type Metadata interface {
	Metadata(eval.Context) (typeSet eval.TypeSet, definitions []Definition)
}
