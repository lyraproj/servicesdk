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
		Active bool `puppet:"name=>import"`
	}

	c := pcore.RootContext()

	// Create a TypeSet from a list of Go structs
	typeSet := c.Reflector().TypeSetFromReflect(`My::Own`, semver.MustParseVersion(`1.0.0`), nil,
		reflect.TypeOf(&Address{}), reflect.TypeOf(&Person{}), reflect.TypeOf(&ExtendedPerson{}))

	// Make the types known to the current loader
	px.AddTypes(c, typeSet)

	tmpDir, err := ioutil.TempDir("", "tsgen_")
	if err == nil {
		//noinspection GoUnhandledErrorResult
		defer os.RemoveAll(tmpDir)
		g := typegen.GetGenerator(`typescript`)
		g.GenerateTypes(typeSet, tmpDir)

		content, err := ioutil.ReadFile(filepath.Join(tmpDir, "My", "Own.ts"))
		if err == nil {
			fmt.Println(string(content))
		}
	}

	if err != nil {
		fmt.Println(err.Error())
	}

	// Output:
	// // this file is generated
	// import {PcoreValue, Value} from 'lyra-workflow';
	//
	// export class Address implements PcoreValue {
	//   readonly street: string;
	//   readonly zipCode: string;
	//
	//   constructor({
	//     street,
	//     zipCode
	//   }: {
	//     street: string,
	//     zipCode: string
	//   }) {
	//     this.street = street;
	//     this.zipCode = zipCode;
	//   }
	//
	//   __pvalue(): {[s: string]: Value} {
	//     const ih: {[s: string]: Value} = {};
	//     ih['street'] = this.street;
	//     ih['zipCode'] = this.zipCode;
	//     return ih;
	//   }
	//
	//   __ptype(): string {
	//     return 'My::Own::Address';
	//   }
	// }
	//
	// export class Person implements PcoreValue {
	//   readonly name: string;
	//   readonly gender: 'male'|'female'|'other';
	//   readonly address: Address|null;
	//
	//   constructor({
	//     name,
	//     gender,
	//     address = null
	//   }: {
	//     name: string,
	//     gender: 'male'|'female'|'other',
	//     address?: Address|null
	//   }) {
	//     this.name = name;
	//     this.gender = gender;
	//     this.address = address;
	//   }
	//
	//   __pvalue(): {[s: string]: Value} {
	//     const ih: {[s: string]: Value} = {};
	//     ih['name'] = this.name;
	//     ih['gender'] = this.gender;
	//     if (this.address !== null) {
	//       ih['address'] = this.address;
	//     }
	//     return ih;
	//   }
	//
	//   __ptype(): string {
	//     return 'My::Own::Person';
	//   }
	// }
	//
	// export class ExtendedPerson extends Person {
	//   readonly import_: boolean;
	//   readonly age: number|null;
	//
	//   constructor({
	//     name,
	//     gender,
	//     import_,
	//     address = null,
	//     age = null
	//   }: {
	//     name: string,
	//     gender: 'male'|'female'|'other',
	//     import_: boolean,
	//     address?: Address|null,
	//     age?: number|null
	//   }) {
	//     super({name: name, gender: gender, address: address});
	//     this.import_ = import_;
	//     this.age = age;
	//   }
	//
	//   __pvalue(): {[s: string]: Value} {
	//     const ih = super.__pvalue();
	//     ih['import'] = this.import_;
	//     if (this.age !== null) {
	//       ih['age'] = this.age;
	//     }
	//     return ih;
	//   }
	//
	//   __ptype(): string {
	//     return 'My::Own::ExtendedPerson';
	//   }
	// }
}
