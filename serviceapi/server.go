package serviceapi

type Service interface {
	Invokable
	Metadata
	StateResolver
}
