package serviceapi

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

var DefinitionMetaType px.Type

type Definition interface {
	px.Value
	issue.Labeled

	// Identifier returns a TypedName that uniquely identifies the activity within the service.
	Identifier() px.TypedName

	// ServiceId is the identifier of the service
	ServiceId() px.TypedName

	// Properties is an ordered map of properties of this definition. Will be of type
	// Hash[Pattern[/\A[a-z][A-Za-z]+\z/],RichData]
	Properties() px.OrderedMap
}

var NewDefinition func(identity, serviceId px.TypedName, properties px.OrderedMap) Definition
