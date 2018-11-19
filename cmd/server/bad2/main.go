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

		sb.RegisterAPI("Foo::Bar", &resource.Bar{})

		s := sb.Server()
		grpc.Serve(c, s)
	})
}
