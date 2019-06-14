package serviceapi

import "github.com/lyraproj/pcore/px"

// Identity defines the API for services that track mappings between internal and external IDs
type Identity interface {
	BumpEra() error
	ReadEra() (era int64, err error)
	AddReference(internalId, otherId string) error
	Associate(internalID string, externalID string) error
	GetExternal(internalID string) (string, error)
	GetInternal(externalID string) (string, error)
	PurgeExternal(externalID string) error
	PurgeInternal(internalID string) error
	PurgeReferences(internalIDPrefix string) error
	RemoveExternal(externalID string) error
	RemoveInternal(internalID string) error
	Search(internalIDPrefix string) (px.List, error)
	Sweep(internalIDPrefix string) error
	Garbage(internalIDPrefix string) (px.List, error)
}
