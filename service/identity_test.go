package service_test

import (
	"fmt"
	"os"

	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-servicesdk/service"
	"github.com/puppetlabs/go-servicesdk/serviceapi"

	// Initialize pcore
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

type identity struct {
}

func (*identity) Associate(internalID, externalID string) error {
	return nil
}

func (*identity) GetExternal(internalID string) (externalID string, ok bool, err error) {
	externalID = "externalID123"
	ok = true
	return
}

func (*identity) GetInternal(externalID string) (internalID string, ok bool, err error) {
	externalID = "internalID456"
	ok = true
	return
}

func (*identity) RemoveExternal(externalID string) error {
	return nil
}

func (*identity) RemoveInternal(internalID string) error {
	return nil
}

func ExampleServerBuilder_RegisterAPI_identity() {
	eval.Puppet.Do(func(c eval.Context) {
		var api serviceapi.Identity
		api = &identity{}
		sb := service.NewServerBuilder(c, `My::Identity::Service`)
		sb.RegisterAPI(serviceapi.IdentityName, api)
		s := sb.Server()
		ts, defs := s.Metadata(c)
		ts.ToString(os.Stdout, eval.PRETTY_EXPANDED, nil)
		fmt.Println()
		fmt.Println(defs)
	})

	// Output:
	// TypeSet[{
	//   pcore_uri => 'http://puppet.com/2016.1/pcore',
	//   pcore_version => '1.0.0',
	//   name_authority => 'http://puppet.com/2016.1/runtime',
	//   name => 'Lyra',
	//   version => '0.1.0',
	//   types => {
	//     Identity => {
	//       functions => {
	//         'associate' => Callable[String, String],
	//         'get_external' => Callable[
	//           [String],
	//           Tuple
	//           [String, Boolean]],
	//         'get_internal' => Callable[
	//           [String],
	//           Tuple
	//           [String, Boolean]],
	//         'remove_external' => Callable[String],
	//         'remove_internal' => Callable[String]
	//       }
	//     }
	//   }
	// }]
	// [Service::Definition('identifier' => TypedName('namespace' => 'definition', 'name' => 'Lyra::Identity'), 'serviceId' => TypedName('namespace' => 'service', 'name' => 'My::Identity::Service'), 'properties' => {'interface' => Lyra::Identity, 'style' => 'callable'})]
}
