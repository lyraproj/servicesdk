package serviceapi

import "github.com/lyraproj/pcore/px"

type Service interface {
	Invokable
	Metadata
	StateResolver

	Identifier(px.Context) px.TypedName
}
