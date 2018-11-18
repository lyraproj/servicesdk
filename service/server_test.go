package service_test

import (
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-servicesdk/service"
	"github.com/puppetlabs/go-servicesdk/wfapi"
	"os"

	// Initialize pcore
	_ "github.com/puppetlabs/go-evaluator/pcore"
	_ "github.com/puppetlabs/go-servicesdk/wf"
)

type testAPI struct{}

func (*testAPI) First(args eval.OrderedMap) string {
	return `first`
}

func (*testAPI) Second(args eval.OrderedMap) string {
	return `second ` + args.Get5(`suffix`, types.WrapString(`nothing`)).String()
}

func ExampleServer_Invoke() {
	eval.Puppet.Do(func(c eval.Context) {
		api := `My::TheApi`
		sb := service.NewServerBuilder(c, `My::Service`)
		sb.RegisterAPI(api, &testAPI{})
		s := sb.Server()

		fmt.Println(s.Invoke(api, `first`, eval.EMPTY_MAP))
		fmt.Println(s.Invoke(api, `second`, eval.Wrap(c, map[string]string{`suffix`: `place`})))
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
		api := `My::TheApi`
		sb := service.NewServerBuilder(c, `My::Service`)
		sb.RegisterAPI(api, &testAPI{})
		sb.RegisterTypes("My", &MyRes{})

		s := sb.Server()
		ts, _ := s.Metadata()

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
	//           [Hash],
	//           String],
	//         'second' => Callable[
	//           [Hash],
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
		_, defs := s.Metadata()
		for _, def := range defs {
			fmt.Println(eval.ToPrettyString(def))
		}
	})

	// Output:
	// Service::Definition(
	//   'identifier' => TypedName(
	//     'namespace' => 'activity',
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
	//           'namespace' => 'activity',
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
		fmt.Println(eval.ToPrettyString(s.State(`My::Test::X`, eval.EMPTY_MAP)))
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
		ts, defs := s.Metadata()
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
	//
}
