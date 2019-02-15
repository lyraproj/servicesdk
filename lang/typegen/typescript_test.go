package typegen

import (
	"bytes"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/semver/semver"
	"io/ioutil"
	"reflect"
	"testing"

	// Initialize pcore
	"fmt"
	_ "github.com/lyraproj/puppet-evaluator/pcore"
	_ "github.com/lyraproj/servicesdk/annotation"
)

func TestGetAllNestedTypes(t *testing.T) {
	eval.Puppet.Do(func(c eval.Context) {
		typesFile := "testdata/aws.pp"
		content, err := ioutil.ReadFile(typesFile)
		if err != nil {
			panic(err.Error())
		}
		ast := c.ParseAndValidate(typesFile, string(content), false)
		c.AddDefinitions(ast)
		_, err = eval.TopEvaluate(c, ast)
		if err != nil {
			panic(err.Error())
		}

		var l interface{}
		var ts eval.TypeSet
		var ok bool

		if l, ok = eval.Load(c, eval.NewTypedName(eval.NsType, `Aws`)); ok {
			ts, ok = l.(eval.TypeSet)
		}
		if !ok {
			panic("Failed to load Aws TypeSet")
		}

		bld := bytes.NewBufferString(``)
		g := NewTsGenerator(c)
		g.GenerateTypes(ts, []string{}, 0, bld)
		fmt.Println(bld.String())
	})
}

func ExampleGenerator_GenerateTypes() {
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

	c := eval.Puppet.RootContext()

	// Create a TypeSet from a list of Go structs
	typeSet := c.Reflector().TypeSetFromReflect(`My::Own`, semver.MustParseVersion(`1.0.0`), nil,
		reflect.TypeOf(&Address{}), reflect.TypeOf(&Person{}), reflect.TypeOf(&ExtendedPerson{}))

	// Make the types known to the current loader
	c.AddTypes(typeSet)

	bld := bytes.NewBufferString(``)
	g := NewTsGenerator(c)
	g.GenerateTypes(typeSet, []string{}, 0, bld)
	fmt.Println(bld.String())

	// Output:
	// export namespace My {
	//   export namespace Own {
	//
	//     export class Address implements PcoreValue {
	//       readonly street: string;
	//       readonly zipCode: string;
	//
	//       constructor({
	//         street,
	//         zipCode
	//       }: {
	//         street: string,
	//         zipCode: string
	//       }) {
	//         this.street = street;
	//         this.zipCode = zipCode;
	//       }
	//
	//       __pvalue(): {[s: string]: Value} {
	//         const ih: {[s: string]: Value} = {};
	//         ih['street'] = this.street;
	//         ih['zipCode'] = this.zipCode;
	//         return ih;
	//       }
	//
	//       __ptype(): string {
	//         return 'My::Own::Address';
	//       }
	//     }
	//
	//     export class Person implements PcoreValue {
	//       readonly name: string;
	//       readonly gender: 'male'|'female'|'other';
	//       readonly address: Address|null;
	//
	//       constructor({
	//         name,
	//         gender,
	//         address = null
	//       }: {
	//         name: string,
	//         gender: 'male'|'female'|'other',
	//         address?: Address|null
	//       }) {
	//         this.name = name;
	//         this.gender = gender;
	//         this.address = address;
	//       }
	//
	//       __pvalue(): {[s: string]: Value} {
	//         const ih: {[s: string]: Value} = {};
	//         ih['name'] = this.name;
	//         ih['gender'] = this.gender;
	//         if (this.address !== null) {
	//           ih['address'] = this.address;
	//         }
	//         return ih;
	//       }
	//
	//       __ptype(): string {
	//         return 'My::Own::Person';
	//       }
	//     }
	//
	//     export class ExtendedPerson extends Person {
	//       readonly enabled: boolean;
	//       readonly age: number|null;
	//
	//       constructor({
	//         name,
	//         gender,
	//         enabled,
	//         address = null,
	//         age = null
	//       }: {
	//         name: string,
	//         gender: 'male'|'female'|'other',
	//         enabled: boolean,
	//         address?: Address|null,
	//         age?: number|null
	//       }) {
	//         super({name: name, gender: gender, address: address});
	//         this.enabled = enabled;
	//         this.age = age;
	//       }
	//
	//       __pvalue(): {[s: string]: Value} {
	//         const ih = super.__pvalue();
	//         ih['enabled'] = this.enabled;
	//         if (this.age !== null) {
	//           ih['age'] = this.age;
	//         }
	//         return ih;
	//       }
	//
	//       __ptype(): string {
	//         return 'My::Own::ExtendedPerson';
	//       }
	//     }
	//   }
	// }
}
