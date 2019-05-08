package service

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/serviceapi"
)

func init() {
	serviceapi.NotFound = func(typeName, extId string) error {
		return px.Error(NotFound, issue.H{`typeName`: typeName, `extId`: extId})
	}
}
