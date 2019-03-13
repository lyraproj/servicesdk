package typegen_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/semver/semver"
	"github.com/lyraproj/servicesdk/lang/typegen"
)

func ExampleGenerator_GenerateTypes_puppet() {
	type Address struct {
		Street  string
		ZipCode string
	}
	type Person struct {
		Name    string
		Gender  string `puppet:"type=>Enum[male,female,other]"`
		Address *Address
	}
	type ExtendedPerson struct {
		Person
		Age    *int `puppet:"type=>Optional[Integer],value=>undef"`
		Active bool `puppet:"name=>enabled"`
	}

	c := pcore.RootContext()

	// Create a TypeSet from a list of Go structs
	typeSet := c.Reflector().TypeSetFromReflect(`My::Own`, semver.MustParseVersion(`1.0.0`), nil,
		reflect.TypeOf(&Address{}), reflect.TypeOf(&Person{}), reflect.TypeOf(&ExtendedPerson{}))

	// Make the types known to the current loader
	px.AddTypes(c, typeSet)

	tmpDir, err := ioutil.TempDir(``, `puppetgen_`)
	if err == nil {
		//noinspection GoUnhandledErrorResult
		defer os.RemoveAll(tmpDir)
		g := typegen.GetGenerator(`puppet`)
		g.GenerateTypes(typeSet, tmpDir)

		content, err := ioutil.ReadFile(filepath.Join(tmpDir, "My", "Own.pp"))
		if err == nil {
			fmt.Println(string(content))
		}
	}

	if err != nil {
		fmt.Println(err.Error())
	}

	// Output:
	// # this file is generated
	// type My::Own = TypeSet[{
	//   pcore_uri => 'http://puppet.com/2016.1/pcore',
	//   pcore_version => '1.0.0',
	//   name_authority => 'http://puppet.com/2016.1/runtime',
	//   name => 'My::Own',
	//   version => '1.0.0',
	//   types => {
	//     Address => {
	//       attributes => {
	//         'street' => String,
	//         'zipCode' => String
	//       }
	//     },
	//     Person => {
	//       attributes => {
	//         'name' => String,
	//         'gender' => Enum['male', 'female', 'other'],
	//         'address' => {
	//           'type' => Optional[Address],
	//           'value' => undef
	//         }
	//       }
	//     },
	//     ExtendedPerson => Person{
	//       attributes => {
	//         'age' => {
	//           'type' => Optional[Integer],
	//           'value' => undef
	//         },
	//         'enabled' => Boolean
	//       }
	//     }
	//   }
	// }]
}
