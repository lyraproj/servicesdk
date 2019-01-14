package serviceapi

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
)

var Definition_Type eval.Type

type Definition interface {
	eval.Value
	issue.Labeled

	// Identifier returns a TypedName that uniquely identifies the activity within the service.
	Identifier() eval.TypedName

	// ServiceId is the identifier of the service
	ServiceId() eval.TypedName

	// Properties is an ordered map of properties of this definition. Will be of type
	// Hash[Pattern[/\A[a-z][A-Za-z]+\z/],RichData]
	Properties() eval.OrderedMap
}

var NewDefinition func(identity, serviceId eval.TypedName, properties eval.OrderedMap) Definition
