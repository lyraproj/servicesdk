package serviceapi

import (
	"github.com/puppetlabs/go-evaluator/eval"
)

// Identifier TypedName namespaces. Used by a service to identify what the type of entity to look for.

// Interface denotes an entity that must have an "interface" property that appoints
// an object type which in turn contains a declaration of the methods that the interface
// implements.
const NsInterface = eval.Namespace(`interface`)

// Activity denotes an entity that can participate in a workflow. The entity must
// declare input and output parameters. An activity of type "action" may also be an interface
// in which case it must have an "interface" property
const NsActivity = eval.Namespace(`activity`)

// ServiceId TypedName namespaces. Used by the Loader to determine the right type
// of RPC mechanism to use when communicating with the service.

// Plugin denotes a service that is a Hashicorp go-plugin
const NsService = eval.Namespace(`service`)

// RESTFul denotes a service that is a RESTFul http or https service.
const RESTFul = eval.Namespace(`RESTFul`)

type Definition interface {
	eval.Value

	// Identifier returns a TypedName that uniquely identifies the activity within the service.
	Identifier() eval.TypedName

	// ServiceId is the identifier of the service
	ServiceId() eval.TypedName

	// Properties is an ordered map of properties of this definition. Will be of type
	// Hash[Pattern[/\A[a-z][A-Za-z]+\z/],RichData]
	Properties() eval.OrderedMap
}

var NewDefinition func(identity, serviceId eval.TypedName, properties eval.OrderedMap) Definition
