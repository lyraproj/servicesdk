package service_test

import (
	"fmt"
	"reflect"

	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/service"
)

func ExampleErrorMetaType() {
	type TestStruct struct {
		Message   string
		Kind      string
		IssueCode string `puppet:"name => issue_code"`
	}

	c := pcore.RootContext()
	ts := &TestStruct{`the message`, `THE_KIND`, `THE_CODE`}
	et, _ := px.Load(c, px.NewTypedName(px.NsType, `Error`))
	ev := et.(px.ObjectType).FromReflectedValue(c, reflect.ValueOf(ts).Elem())
	fmt.Println(service.ErrorMetaType.IsInstance(ev, nil))
	// Output: true
}
