package wf

import (
	"fmt"

	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
)

func ExampleParse() {
	pcore.Do(func(px.Context) {
		c := Parse("hello")
		fmt.Println(c)
	})
	// Output: hello
}

func ExampleParse_and() {
	pcore.Do(func(px.Context) {
		c := Parse("hello and goodbye")
		fmt.Println(c)
	})
	// Output: hello and goodbye
}

func ExampleParse_not() {
	pcore.Do(func(px.Context) {
		c := Parse("!(hello and goodbye)")
		fmt.Println(c)
	})
	// Output: !(hello and goodbye)
}

func ExampleParse_or() {
	pcore.Do(func(px.Context) {
		c := Parse("greeting and (hello or goodbye)")
		fmt.Println(c)
	})
	// Output: greeting and (hello or goodbye)
}
