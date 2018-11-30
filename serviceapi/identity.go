package serviceapi

// IdentityName is the name associated with identity services by the loader
var IdentityName = "Lyra::Identity"

// Identity defines the API for services that track mappings between internal and external IDs
type Identity interface {
	Associate(internalID, externalID string)
	GetExternal(internalID string) (externalID string)
	GetInternal(externalID string) (internalID string)
	RemoveExternal(externalID string)
	RemoveInternal(internalID string)
}
