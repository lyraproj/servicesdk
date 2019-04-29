package typegen_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/semver/semver"
	"github.com/lyraproj/servicesdk/lang/typegen"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GenerateTypes_go(t *testing.T) {
	type Address struct {
		Street  string
		ZipCode string
	}
	type Person struct {
		Name    string
		Gender  string
		Address *Address
	}
	type ExtendedPerson struct {
		Person
		Age    *int
		Active bool `puppet:"name=>import"`
	}

	expected := `// this file is generated
package own

import (
	"fmt"
	"reflect"

	"github.com/lyraproj/pcore/px"
)

type Address struct {
	Street  string
	ZipCode string
}

type Person struct {
	Name    string
	Gender  string
	Address *Address
}

type ExtendedPerson struct {
	Name    string
	Gender  string
	Import  bool
	Address *Address
	Age     *int64
}

func InitTypes(c px.Context) {
	load := func(n string) px.Type {
		if v, ok := px.Load(c, px.NewTypedName(px.NsType, n)); ok {
			return v.(px.Type)
		}
		panic(fmt.Errorf("unable to load Type '%s'", n))
	}

	ir := c.ImplementationRegistry()
	ir.RegisterType(load("My::Own::Address"), reflect.TypeOf(&Address{}))
	ir.RegisterType(load("My::Own::ExtendedPerson"), reflect.TypeOf(&ExtendedPerson{}))
	ir.RegisterType(load("My::Own::Person"), reflect.TypeOf(&Person{}))
}
`
	pcore.Do(func(c px.Context) {
		// Create a TypeSet from a list of Go structs
		typeSet := c.Reflector().TypeSetFromReflect(`My::Own`, semver.MustParseVersion(`1.0.0`), nil,
			reflect.TypeOf(&Address{}), reflect.TypeOf(&Person{}), reflect.TypeOf(&ExtendedPerson{}))

		// Make the types known to the current loader
		px.AddTypes(c, typeSet)

		tmpDir, err := ioutil.TempDir("", "gogen_")

		if err == nil {
			//noinspection GoUnhandledErrorResult
			defer os.RemoveAll(tmpDir)
			g := typegen.GetGenerator(`go`)
			g.GenerateTypes(typeSet, tmpDir)

			var content []byte
			content, err = ioutil.ReadFile(filepath.Join(tmpDir, "my", "own", "own.go"))
			if err == nil {
				require.Equal(t, expected, string(content))
			}
		}

		if err != nil {
			t.Error(err)
		}
	})
}
