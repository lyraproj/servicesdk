package service_test

import (
	"fmt"
	"github.com/puppetlabs/go-servicesdk/service"
)

func ExampleFindCommonNamespace() {
	fmt.Println(service.FindCommonNamespace([]string{
		"My::Common::A::A",
		"My::Common::A::B",
		"My::Common::B",
	}))
	fmt.Println(service.FindCommonNamespace([]string{
		"My::Common::A::A",
		"My::Common::A::B",
		"My::Other::B",
	}))
	fmt.Println(service.FindCommonNamespace([]string{
		"My::Common::A::A",
		"Your::Common::A::B",
		"My::Common::B",
	}))

	// Output:
	// My::Common true
	// My true
	//  false
}
