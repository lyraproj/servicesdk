package serviceapi

import (
	"github.com/lyraproj/puppet-evaluator/eval"
)

type Metadata interface {
	Metadata(eval.Context) (typeSet eval.TypeSet, definitions []Definition)
}
