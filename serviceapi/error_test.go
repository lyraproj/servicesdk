package serviceapi_test

import (
	"fmt"
	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/servicesdk/serviceapi"
	"reflect"

	_ "github.com/lyraproj/servicesdk/service"
)

func ExampleNewError_reflectTo() {
	type TestStruct struct {
		Message   string
		Kind      string
		IssueCode string `puppet:"name => issue_code"`
	}

	c := pcore.RootContext()
	ts := &TestStruct{}

	ev := serviceapi.NewError(c, `the message`, `THE_KIND`, `THE_CODE`, nil, nil)
	c.Reflector().ReflectTo(ev, reflect.ValueOf(ts).Elem())
	fmt.Printf("\nmessage: %s, kind %s, issueCode %s\n", ts.Message, ts.Kind, ts.IssueCode)
	// Output: message: the message, kind THE_KIND, issueCode THE_CODE
}
