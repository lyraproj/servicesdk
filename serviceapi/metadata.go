package serviceapi

import (
	"github.com/lyraproj/pcore/px"
)

type Metadata interface {
	Metadata(px.Context) (typeSet px.TypeSet, definitions []Definition)
}
