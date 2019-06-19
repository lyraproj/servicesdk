package serviceapi

import "github.com/lyraproj/pcore/px"

// Identity defines the API for services that track mappings between internal and external IDs
type Identity interface {
	BumpEra(px.Context)
	AddReference(ctx px.Context, internalId, otherId string)
	Associate(ctx px.Context, internalID string, externalID string)
	GetExternal(ctx px.Context, internalID string) (string, bool)
	GetInternal(ctx px.Context, externalID string) (string, bool)
	PurgeExternal(ctx px.Context, externalID string)
	PurgeInternal(ctx px.Context, internalID string)
	PurgeReferences(ctx px.Context, internalIDPrefix string)
	RemoveExternal(ctx px.Context, externalID string)
	RemoveInternal(ctx px.Context, internalID string)
	Search(ctx px.Context, internalIDPrefix string) px.List
	Sweep(ctx px.Context, internalIDPrefix string)
	Garbage(ctx px.Context, internalIDPrefix string) px.List
}
