package serviceapi

import "github.com/lyraproj/pcore/px"

// Identity defines the API for services that track mappings between internal and external IDs
type Identity interface {
	BumpEra()
	ReadEra() int64
	AddReference(internalId, otherId string)
	Associate(internalID string, externalID string)
	GetExternal(internalID string) string
	GetInternal(externalID string) string
	PurgeExternal(externalID string)
	PurgeInternal(internalID string)
	PurgeReferences(internalIDPrefix string)
	RemoveExternal(externalID string)
	RemoveInternal(internalID string)
	Search(internalIDPrefix string) px.List
	Sweep(internalIDPrefix string)
	Garbage(internalIDPrefix string) px.List
}
