package service_test

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/lyraproj/servicesdk/lang/go/lyra"

	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/serialization"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/annotation"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/wf"
)

type testAPI struct{}

func (*testAPI) First() string {
	return `first`
}

func (*testAPI) Second(suffix string) string {
	return `second ` + suffix
}

func ExampleServer_Invoke() {
	pcore.Do(func(c px.Context) {
		api := `My::TheApi`
		sb := service.NewServiceBuilder(c, `My::Service`)

		sb.RegisterAPI(api, &testAPI{})

		s := sb.Server()
		fmt.Println(s.Invoke(c, api, `first`))
		fmt.Println(s.Invoke(c, api, `second`, px.Wrap(c, `place`)))
	})

	// Output:
	// first
	// second place
}

type MyRes struct {
	Name  string
	Phone string
}

func ExampleServer_Metadata_typeSet() {
	pcore.Do(func(c px.Context) {
		sb := service.NewServiceBuilder(c, `My::Service`)

		sb.RegisterAPI(`My::TheApi`, &testAPI{})
		sb.RegisterTypes("My", &MyRes{})

		s := sb.Server()
		ts, _ := s.Metadata(c)
		ts.ToString(os.Stdout, px.PrettyExpanded, nil)
		fmt.Println()
	})

	// Output:
	// TypeSet[{
	//   pcore_uri => 'http://puppet.com/2016.1/pcore',
	//   pcore_version => '1.0.0',
	//   name_authority => 'http://puppet.com/2016.1/runtime',
	//   name => 'My',
	//   version => '0.1.0',
	//   types => {
	//     MyRes => {
	//       attributes => {
	//         'name' => String,
	//         'phone' => String
	//       }
	//     },
	//     TheApi => {
	//       functions => {
	//         'first' => Callable[
	//           [0, 0],
	//           String],
	//         'second' => Callable[
	//           [String],
	//           String]
	//       }
	//     }
	//   }
	// }]
}

type MyOuterRes struct {
	Who  *MyRes
	What string
}

func ExampleBuilder_RegisterTypes_nestedType() {
	pcore.Do(func(c px.Context) {
		sb := service.NewServiceBuilder(c, `My::Service`)

		sb.RegisterTypes("My", &MyOuterRes{})

		s := sb.Server()
		ts, _ := s.Metadata(c)
		ts.ToString(os.Stdout, px.PrettyExpanded, nil)
		fmt.Println()
	})

	// Output:
	// TypeSet[{
	//   pcore_uri => 'http://puppet.com/2016.1/pcore',
	//   pcore_version => '1.0.0',
	//   name_authority => 'http://puppet.com/2016.1/runtime',
	//   name => 'My',
	//   version => '0.1.0',
	//   types => {
	//     MyOuterRes => {
	//       attributes => {
	//         'who' => Optional[MyRes],
	//         'what' => String
	//       }
	//     },
	//     MyRes => {
	//       attributes => {
	//         'name' => String,
	//         'phone' => String
	//       }
	//     }
	//   }
	// }]
}

type Person struct {
	Who      *MyRes
	Children []*Person
	Born     time.Time
}

func ExampleBuilder_RegisterTypes_recursiveType() {
	pcore.Do(func(c px.Context) {
		sb := service.NewServiceBuilder(c, `My::Service`)

		sb.RegisterTypes("My", &Person{})

		s := sb.Server()
		ts, _ := s.Metadata(c)
		ts.ToString(os.Stdout, px.PrettyExpanded, nil)
		fmt.Println()
	})

	// Output:
	// TypeSet[{
	//   pcore_uri => 'http://puppet.com/2016.1/pcore',
	//   pcore_version => '1.0.0',
	//   name_authority => 'http://puppet.com/2016.1/runtime',
	//   name => 'My',
	//   version => '0.1.0',
	//   types => {
	//     MyRes => {
	//       attributes => {
	//         'name' => String,
	//         'phone' => String
	//       }
	//     },
	//     Person => {
	//       attributes => {
	//         'who' => Optional[MyRes],
	//         'children' => Array[Optional[Person]],
	//         'born' => Timestamp
	//       }
	//     }
	//   }
	// }]
}

func ExampleServer_Metadata_definitions() {
	pcore.Do(func(c px.Context) {
		sb := service.NewServiceBuilder(c, `My::Service`)

		sb.RegisterTypes("My", &MyRes{})
		sb.RegisterActivity((&lyra.Workflow{
			Name: `My::Test`,
			Activities: []lyra.Activity{
				&lyra.Resource{
					Name: `X`,
					State: func(struct {
						A string
						B string
					}) *MyRes {
						return &MyRes{Name: `Bob`, Phone: `12345`}
					}}}}).Resolve(c, ``))

		s := sb.Server()
		_, defs := s.Metadata(c)
		for _, def := range defs {
			fmt.Println(px.ToPrettyString(def))
		}
	})

	// Output:
	// Service::Definition(
	//   'identifier' => TypedName(
	//     'namespace' => 'definition',
	//     'name' => 'My::Test'
	//   ),
	//   'serviceId' => TypedName(
	//     'namespace' => 'service',
	//     'name' => 'My::Service'
	//   ),
	//   'properties' => {
	//     'activities' => [
	//       Service::Definition(
	//         'identifier' => TypedName(
	//           'namespace' => 'definition',
	//           'name' => 'My::Test::X'
	//         ),
	//         'serviceId' => TypedName(
	//           'namespace' => 'service',
	//           'name' => 'My::Service'
	//         ),
	//         'properties' => {
	//           'input' => [
	//             Parameter(
	//               'name' => 'a',
	//               'type' => String
	//             ),
	//             Parameter(
	//               'name' => 'b',
	//               'type' => String
	//             )],
	//           'resourceType' => My::MyRes,
	//           'style' => 'resource'
	//         }
	//       )],
	//     'style' => 'workflow'
	//   }
	// )
	//
}

type OwnerRes struct {
	Id    *string
	Phone string
}

type ContainedRes struct {
	Id      *string
	OwnerId string
	Stuff   string
}

func ExampleBuilder_RegisterTypes_annotatedTypeSet() {
	pcore.Do(func(c px.Context) {
		sb := service.NewServiceBuilder(c, `My::Service`)

		sb.RegisterTypes("My",
			sb.BuildResource(&OwnerRes{}, func(rtb service.ResourceTypeBuilder) {
				rtb.ProvidedAttributes(`id`)
				rtb.ImmutableAttributes(`telephoneNumber`)
				rtb.Tags(map[string]string{`Phone`: `name=>telephoneNumber`})
				rtb.AddRelationship(`mine`, `My::ContainedRes`, annotation.KindContained, annotation.CardinalityMany, ``, []string{`id`, `ownerId`})
			}),
			sb.BuildResource(&ContainedRes{}, func(rtb service.ResourceTypeBuilder) {
				rtb.ProvidedAttributes(`id`)
				rtb.AddRelationship(`owner`, `My::OwnerRes`, annotation.KindContainer, annotation.CardinalityOne, ``, []string{`ownerId`, `id`})
			}),
		)
		s := sb.Server()
		ts, md := s.Metadata(c)
		bld := bytes.NewBufferString(``)
		coll := serialization.NewJsonStreamer(bld)

		sr := serialization.NewSerializer(pcore.RootContext(), px.EmptyMap)
		sr.Convert(types.WrapValues([]px.Value{ts, px.Wrap(c, md)}), coll)

		dr := serialization.NewDeserializer(c, px.EmptyMap)
		serialization.JsonToData(`/tmp/tst`, bld, dr)
		dt := dr.Value().(*types.Array)
		dt.At(0).ToString(os.Stdout, px.PrettyExpanded, nil)
		fmt.Println()
	})

	// Output:
	// TypeSet[{
	//   pcore_uri => 'http://puppet.com/2016.1/pcore',
	//   pcore_version => '1.0.0',
	//   name_authority => 'http://puppet.com/2016.1/runtime',
	//   name => 'My',
	//   version => '0.1.0',
	//   types => {
	//     ContainedRes => {
	//       annotations => {
	//         Lyra::Resource => {
	//           'providedAttributes' => ['id'],
	//           'relationships' => {
	//             'owner' => {
	//               'type' => OwnerRes,
	//               'kind' => 'container',
	//               'cardinality' => 'one',
	//               'keys' => ['ownerId', 'id']
	//             }
	//           }
	//         }
	//       },
	//       attributes => {
	//         'id' => Optional[String],
	//         'ownerId' => String,
	//         'stuff' => String
	//       }
	//     },
	//     OwnerRes => {
	//       annotations => {
	//         Lyra::Resource => {
	//           'immutableAttributes' => ['telephoneNumber'],
	//           'providedAttributes' => ['id'],
	//           'relationships' => {
	//             'mine' => {
	//               'type' => ContainedRes,
	//               'kind' => 'contained',
	//               'cardinality' => 'many',
	//               'keys' => ['id', 'ownerId']
	//             }
	//           }
	//         }
	//       },
	//       attributes => {
	//         'id' => Optional[String],
	//         'telephoneNumber' => String
	//       }
	//     }
	//   }
	// }]
	//
}

func ExampleServer_Metadata_state() {
	pcore.Do(func(c px.Context) {
		sb := service.NewServiceBuilder(c, `My::Service`)

		sb.RegisterTypes("My", &MyRes{})
		sb.RegisterStateConverter(lyra.StateConverter)
		sb.RegisterActivity((&lyra.Workflow{
			Name: `My::Test`,
			Activities: []lyra.Activity{
				&lyra.Resource{
					Name: `X`,
					State: func(input struct {
						A string
						B string
					}) *MyRes {
						return &MyRes{Name: `Bob`, Phone: `12345`}
					}}}}).Resolve(c, ``))

		s := sb.Server()
		fmt.Println(px.ToPrettyString(s.State(c, `My::Test::X`, px.EmptyMap)))
	})

	// Output:
	// My::MyRes(
	//   'name' => 'Bob',
	//   'phone' => '12345'
	// )
}

type MyIdentityService struct {
	extToId map[string]px.URI
	idToExt map[px.URI]string
}

func (is *MyIdentityService) GetExternal(id px.URI) (string, error) {
	if ext, ok := is.idToExt[id]; ok {
		return ext, nil
	}
	return ``, wf.NotFound
}

func (is *MyIdentityService) GetInternal(ext string) (px.URI, error) {
	if id, ok := is.extToId[ext]; ok {
		return id, nil
	}
	return px.URI(``), wf.NotFound
}

func ExampleServer_Metadata_api() {
	pcore.Do(func(c px.Context) {
		sb := service.NewServiceBuilder(c, `My::Service`)

		sb.RegisterAPI(`My::Identity`, &MyIdentityService{map[string]px.URI{}, map[px.URI]string{}})

		s := sb.Server()
		ts, defs := s.Metadata(c)
		ts.ToString(os.Stdout, px.PrettyExpanded, nil)
		fmt.Println()
		for _, def := range defs {
			fmt.Println(px.ToPrettyString(def))
		}
	})

	// Output:
	// TypeSet[{
	//   pcore_uri => 'http://puppet.com/2016.1/pcore',
	//   pcore_version => '1.0.0',
	//   name_authority => 'http://puppet.com/2016.1/runtime',
	//   name => 'My',
	//   version => '0.1.0',
	//   types => {
	//     Identity => {
	//       functions => {
	//         'getExternal' => Callable[
	//           [String],
	//           String],
	//         'getInternal' => Callable[
	//           [String],
	//           String]
	//       }
	//     }
	//   }
	// }]
	// Service::Definition(
	//   'identifier' => TypedName(
	//     'namespace' => 'definition',
	//     'name' => 'My::Identity'
	//   ),
	//   'serviceId' => TypedName(
	//     'namespace' => 'service',
	//     'name' => 'My::Service'
	//   ),
	//   'properties' => {
	//     'interface' => My::Identity,
	//     'style' => 'callable'
	//   }
	// )
	//
}
