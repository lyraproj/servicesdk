package serviceapi

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

type ErrorObject interface {
	px.PuppetObject

	// Kind returns the error kind
	Kind() string

	// Message returns the error message
	Message() string

	// IssueCode returns the issue code
	IssueCode() string

	// PartialResult returns the optional partial result. It returns
	// pcore.UNDEF if no partial result exists
	PartialResult() px.Value

	// Details returns the optional details. It returns
	// an empty map when no details exist
	Details() px.OrderedMap
}

var ErrorFromReported func(c px.Context, err issue.Reported) ErrorObject

var NewError func(c px.Context, message, kind, issueCode string, partialResult px.Value, details px.OrderedMap) ErrorObject

// NotFound returns the special NotFound error which is recognized by the Lyra workflow engine. It should
// be used when requests are made to read, update, or delete a resource with an external id that no longer exists.
var NotFound func(typeName, extId string) error
