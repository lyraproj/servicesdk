package condition

import (
	"fmt"
	"github.com/lyraproj/puppet-evaluator/eval"

	// Ensure pcore initialization
	_ "github.com/lyraproj/puppet-evaluator/pcore"
)

func ExampleParse() {
	eval.Puppet.Do(func(eval.Context) {
		c := Parse("hello")
		fmt.Println(c)
	})
	// Output: hello
}

func ExampleParse_and() {
	eval.Puppet.Do(func(eval.Context) {
		c := Parse("hello and goodbye")
		fmt.Println(c)
	})
	// Output: hello and goodbye
}

func ExampleParse_not() {
	eval.Puppet.Do(func(eval.Context) {
		c := Parse("!(hello and goodbye)")
		fmt.Println(c)
	})
	// Output: !(hello and goodbye)
}

func ExampleParse_or() {
	eval.Puppet.Do(func(eval.Context) {
		c := Parse("greeting and (hello or goodbye)")
		fmt.Println(c)
	})
	// Output: greeting and (hello or goodbye)
}
