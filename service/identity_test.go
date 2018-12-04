package service_test

import (
	"fmt"
	"os"

	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/serviceapi"

	// Initialize pcore
	_ "github.com/lyraproj/puppet-evaluator/pcore"
)

type identity struct {
}

func (*identity) Associate(internalID, externalID string) {
	return
}

func (*identity) GetExternal(internalID string) (externalID string) {
	externalID = "externalID123"
	return
}

func (*identity) GetInternal(externalID string) (internalID string) {
	externalID = "internalID456"
	return
}

func (*identity) RemoveExternal(externalID string) {
}

func (*identity) RemoveInternal(internalID string) {
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
	//         'associate' => Callable[
	//           [String, String],
	//           Any],
	//         'get_external' => Callable[
	//           [String],
	//           String],
	//         'get_internal' => Callable[
	//           [String],
	//           String],
	//         'remove_external' => Callable[
	//           [String],
	//           Any],
	//         'remove_internal' => Callable[
	//           [String],
	//           Any]
	//       }
	//     }
	//   }
	// }]
	// [Service::Definition('identifier' => TypedName('namespace' => 'definition', 'name' => 'Lyra::Identity'), 'serviceId' => TypedName('namespace' => 'service', 'name' => 'My::Identity::Service'), 'properties' => {'interface' => Lyra::Identity, 'style' => 'callable'})]
	//
}
