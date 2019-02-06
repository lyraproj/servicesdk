package lang

import (
	"bytes"
	"github.com/lyraproj/puppet-evaluator/eval"
)

// The Generator interface is implemented by generators that can transform Pcore types
// to types in some specific language.
type Generator interface {
	// GenerateTypes produces types in some language for all types in the given TypeSet and
	// appends them to the given buffer.
	GenerateTypes(ts eval.TypeSet, ns []string, indent int, bld *bytes.Buffer)

	// GenerateType produces a type in some language for the given Type and appends it to
	// the given buffer.
	GenerateType(t eval.Type, ns []string, indent int, bld *bytes.Buffer)
}
