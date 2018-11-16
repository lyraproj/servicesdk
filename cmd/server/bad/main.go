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
		//plugin.go: panic: type/typeset clash
		evs := sb.RegisterTypes("Foo::Foo3", resource.CrdResource{})
		sb.RegisterHandler("Foo::Foo3", &resource.CrdHandler{}, evs[0])

		//option 2: uncomment these 2 lines to see the error:
		//plugin.go: panic: registered types share no common namespace
		// res := eval.Wrap(c, resource.CrdResource{})
		// sb.RegisterHandler("Foo::Foo3", &resource.CrdHandler{}, res.PType())

		s := sb.Server()
		grpc.Serve(s)
	})
}
