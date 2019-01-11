package service_test

import (
	"bytes"
	"fmt"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/serialization"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/annotation"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/wfapi"
	"os"

	// Initialize pcore
	_ "github.com/lyraproj/puppet-evaluator/pcore"
	_ "github.com/lyraproj/servicesdk/wf"
)

type testAPI struct{}

func (*testAPI) First() string {
	return `first`
}

func (*testAPI) Second(suffix string) string {
	return `second ` + suffix
}

func ExampleServer_Invoke() {
	eval.Puppet.Do(func(c eval.Context) {
		api := `My::TheApi`
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterAPI(api, &testAPI{})

		s := sb.Server()
		fmt.Println(s.Invoke(c, api, `first`))
		fmt.Println(s.Invoke(c, api, `second`, eval.Wrap(c, `place`)))
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
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterAPI(`My::TheApi`, &testAPI{})
		sb.RegisterTypes("My", &MyRes{})

		s := sb.Server()
		ts, _ := s.Metadata(c)
		ts.ToString(os.Stdout, eval.PRETTY_EXPANDED, nil)
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

func ExampleServer_Metadata_definitions() {
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterTypes("My", &MyRes{})
		sb.RegisterActivity(wfapi.NewWorkflow(c, func(b wfapi.WorkflowBuilder) {
			b.Name(`My::Test`)
			b.Resource(func(w wfapi.ResourceBuilder) {
				w.Name(`X`)
				w.Input(w.Parameter(`a`, `String`))
				w.Input(w.Parameter(`b`, `String`))
				w.StateStruct(&MyRes{Name: `Bob`, Phone: `12345`})
			})
		}))

		s := sb.Server()
		_, defs := s.Metadata(c)
		for _, def := range defs {
			fmt.Println(eval.ToPrettyString(def))
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
	//           'resource_type' => My::MyRes,
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

func ExampleServer_TypeSet_annotated() {
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterTypes("My",
			sb.BuildResource(&OwnerRes{}, func(rtb service.ResourceTypeBuilder) {
				rtb.ProvidedAttributes(`id`)
				rtb.ImmutableAttributes(`telephone_number`)
				rtb.Tags(map[string]string{`Phone`: `name=>telephone_number`})
				rtb.AddRelationship(`mine`, `My::ContainedRes`, annotation.KindContained, annotation.CardinalityMany, ``, []string{`id`, `owner_id`})
			}),
			sb.BuildResource(&ContainedRes{}, func(rtb service.ResourceTypeBuilder) {
				rtb.ProvidedAttributes(`id`)
				rtb.AddRelationship(`owner`, `My::OwnerRes`, annotation.KindContainer, annotation.CardinalityOne, ``, []string{`owner_id`, `id`})
			}),
		)
		s := sb.Server()
		ts, md := s.Metadata(c)
		bld := bytes.NewBufferString(``)
		coll := serialization.NewJsonStreamer(bld)

		sr := serialization.NewSerializer(eval.Puppet.RootContext(), eval.EMPTY_MAP)
		sr.Convert(types.WrapValues([]eval.Value{ts, eval.Wrap(c, md)}), coll)

		dr := serialization.NewDeserializer(c, eval.EMPTY_MAP)
		serialization.JsonToData(`/tmp/tst`, bld, dr)
		dt := dr.Value().(*types.ArrayValue)
		dt.At(0).ToString(os.Stdout, eval.PRETTY_EXPANDED, nil)
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
	//           'provided_attributes' => ['id'],
	//           'relationships' => {
	//             'owner' => {
	//               'type' => OwnerRes,
	//               'kind' => 'container',
	//               'cardinality' => 'one',
	//               'keys' => ['owner_id', 'id']
	//             }
	//           }
	//         }
	//       },
	//       attributes => {
	//         'id' => {
	//           'type' => Optional[String],
	//           'value' => undef
	//         },
	//         'owner_id' => String,
	//         'stuff' => String
	//       }
	//     },
	//     OwnerRes => {
	//       annotations => {
	//         Lyra::Resource => {
	//           'immutable_attributes' => ['telephone_number'],
	//           'provided_attributes' => ['id'],
	//           'relationships' => {
	//             'mine' => {
	//               'type' => ContainedRes,
	//               'kind' => 'contained',
	//               'cardinality' => 'many',
	//               'keys' => ['id', 'owner_id']
	//             }
	//           }
	//         }
	//       },
	//       attributes => {
	//         'id' => {
	//           'type' => Optional[String],
	//           'value' => undef
	//         },
	//         'telephone_number' => String
	//       }
	//     }
	//   }
	// }]
	//
}

func ExampleServer_Metadata_state() {
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterTypes("My", &MyRes{})
		sb.RegisterStateConverter(service.GoStateConverter)
		sb.RegisterActivity(wfapi.NewWorkflow(c, func(b wfapi.WorkflowBuilder) {
			b.Name(`My::Test`)
			b.Resource(func(w wfapi.ResourceBuilder) {
				w.Name(`X`)
				w.Input(w.Parameter(`a`, `String`))
				w.Input(w.Parameter(`b`, `String`))
				w.StateStruct(&MyRes{Name: `Bob`, Phone: `12345`})
			})
		}))

		s := sb.Server()
		fmt.Println(eval.ToPrettyString(s.State(c, `My::Test::X`, eval.EMPTY_MAP)))
	})

	// Output:
	// My::MyRes(
	//   'name' => 'Bob',
	//   'phone' => '12345'
	// )
}

type MyIdentityService struct {
	extToId map[string]eval.URI
	idToExt map[eval.URI]string
}

func (is *MyIdentityService) GetExternal(id eval.URI) (string, error) {
	if ext, ok := is.idToExt[id]; ok {
		return ext, nil
	}
	return ``, wfapi.NotFound
}

func (is *MyIdentityService) GetInternal(ext string) (eval.URI, error) {
	if id, ok := is.extToId[ext]; ok {
		return id, nil
	}
	return eval.URI(``), wfapi.NotFound
}

func ExampleServer_Metadata_api() {
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterAPI(`My::Identity`, &MyIdentityService{map[string]eval.URI{}, map[eval.URI]string{}})

		s := sb.Server()
		ts, defs := s.Metadata(c)
		ts.ToString(os.Stdout, eval.PRETTY_EXPANDED, nil)
		fmt.Println()
		for _, def := range defs {
			fmt.Println(eval.ToPrettyString(def))
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
	//         'get_external' => Callable[
	//           [String],
	//           String],
	//         'get_internal' => Callable[
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
