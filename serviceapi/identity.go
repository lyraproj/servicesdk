package serviceapi

// IdentityName is the name associated with identity services by the loader
var IdentityName = "Lyra::Identity"

// Identity defines the API for services that track mappings between internal and external IDs
type Identity interface {
	Associate(internalID, externalID string) error
	GetExternal(internalID string) (externalID string, ok bool, err error)
	GetInternal(externalID string) (internalID string, ok bool, err error)
	RemoveExternal(externalID string) error
	RemoveInternal(internalID string) error
}
