package main

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-servicesdk/cmd/server/resource"
	"github.com/puppetlabs/go-servicesdk/grpc"
	"github.com/puppetlabs/go-servicesdk/service"

	// Initialize pcore
	// Ensure that pcore is initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

func main() {
	eval.Puppet.Do(func(c eval.Context) {

		sb := service.NewServerBuilder(c, `Foo`)

		//what is the correct way to register a handler?

		//option 1:
		// TH: This is the correct way. Types must be registered before an
		// attempt is made to register a handler for that type unless the
		// type is previously known to the eval.Context used when creating
		// the builder
		evs := sb.RegisterTypes("Foo::CrdResource", resource.CrdResource{})
		sb.RegisterHandler("Foo::CrdHandler", &resource.CrdHandler{}, evs[0])

		s := sb.Server()
		grpc.Serve(c, s)
	})
}
