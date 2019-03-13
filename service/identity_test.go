package service_test

import (
	"fmt"
	"os"

	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/serviceapi"
)

type identity struct {
}

func (*identity) Associate(internalID, externalID string) {
}

func (*identity) GetExternal(internalID string) (externalID string) {
	externalID = "externalID123"
	return
}

func (*identity) GetInternal(externalID string) (internalID string) {
	internalID = "internalID456"
	return
}

func (*identity) RemoveExternal(externalID string) {
}

func (*identity) RemoveInternal(internalID string) {
}

func ExampleBuilder_RegisterAPI_identity() {
	pcore.Do(func(c px.Context) {
		api := &identity{}
		sb := service.NewServiceBuilder(c, `My::Identity::Service`)
		sb.RegisterAPI(serviceapi.IdentityName, api)
		s := sb.Server()
		ts, defs := s.Metadata(c)
		ts.ToString(os.Stdout, px.PrettyExpanded, nil)
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
	//         'getExternal' => Callable[
	//           [String],
	//           String],
	//         'getInternal' => Callable[
	//           [String],
	//           String],
	//         'removeExternal' => Callable[
	//           [String],
	//           Any],
	//         'removeInternal' => Callable[
	//           [String],
	//           Any]
	//       }
	//     }
	//   }
	// }]
	// [Service::Definition('identifier' => TypedName('namespace' => 'definition', 'name' => 'Lyra::Identity'), 'serviceId' => TypedName('namespace' => 'service', 'name' => 'My::Identity::Service'), 'properties' => {'interface' => Lyra::Identity, 'style' => 'callable'})]
	//
}
