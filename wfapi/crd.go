package wfapi

import "github.com/lyraproj/pcore/px"

type ErrorConstant string

func (e ErrorConstant) Error() string {
	return string(e)
}

// Error returned by Read, Delete, and Update when the requested state isn't found
const NotFound = ErrorConstant(`not found`)

type CRD interface {
	// Create creates the desired state and returns a possibly amended version of that state
	// together with the externalId by which the state can henceforth be identified.
	Create(state px.OrderedMap) (px.OrderedMap, string, error)

	// Read reads and returns the current state identified by the given externalId. The error NotFound
	// is returned when no state can be found.
	Read(externalId string) (px.OrderedMap, error)

	// Delete deletes the state identified by the given externalId. The error NotFound is returned when
	// no state can be found.
	Delete(externalId string) error
}

type CRUD interface {
	CRD

	// Update updates the state identified by the given externalId to a new state and returns a possibly
	// amended version of that state. The error NotFound is returned when no state can be found.
	Update(externalId string, state px.OrderedMap) (px.OrderedMap, error)
}
