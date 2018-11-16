package grpc

import (
	"fmt"
	"os/exec"
	"testing"

	eval "github.com/puppetlabs/go-evaluator/eval"
	types "github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-servicesdk/cmd/server/resource"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/stretchr/testify/require"
)

func TestInvoke(t *testing.T) {
	ik := invokable(t)
	actual := ik.Invoke("Foo::Foo2", "hello", types.WrapString("Cassie"))
	require.Equal(t, "Hello Cassie!", actual.String())
}

func TestInvoke_CanReturnStructType(t *testing.T) {
	// This read should return a structured actual state but returns a string
	ik := invokable(t)
	actual := ik.Invoke("Foo::Foo3", "read", types.WrapString("1234"))
	require.NotEqual(t, "*types.StringValue", fmt.Sprintf("%T", actual))
}

func TestInvoke_FuncReturnsOneArgError(t *testing.T) {
	// "Delete" returns only an error but CallGoReflected expects either:
	// - a single non-error return type, or
	// - a single non-error return type and a single error
	//
	// This call returns no error but causes an index out of bounds
	ik := invokable(t)
	require.NotPanics(t, func() { ik.Invoke("Foo::Foo3", "delete", types.WrapString("1234")) })
}

func TestInvoke_FuncReturnsStringAndError(t *testing.T) {
	// Similar to the example above
	// "Delete2" returns a single non-error return type and a single error
	// which meets the one arg, one error limitation
	//
	// When the error is nil the following occurs: panic: interface conversion: interface is nil, not error
	ik := invokable(t)
	var actual eval.Value
	require.NotPanics(t, func() { actual = ik.Invoke("Foo::Foo3", "delete2", types.WrapString("1234")) })
	fmt.Println(actual)

}

func TestInvoke_FuncReturnsStringAndError_ReturnsActualError(t *testing.T) {
	// Similar to the example above
	// "Delete3" returns a single non-error return type and a single error
	// which meets the one arg, one error limitation
	//
	// When the error is *not* nil it is not treated as an application error,
	// rather the invocation is deemed to have failed. We can't differentiate
	// between a genesis error and a user-application error. We don't propagate
	// the errors through to the client-side caller.
	ik := invokable(t)
	var actual eval.Value
	require.NotPanics(t, func() { actual = ik.Invoke("Foo::Foo3", "delete3", types.WrapString("1234")) })
	fmt.Println(actual)
}

func TestInvoke_CanReceiveStruct(t *testing.T) {
	c := eval.Puppet.RootContext()

	// It would be great to be able to create the eval.Value needed to call create
	// This example serializes the structure to a string and reports this warning:
	// [0] contains a Runtime value. It will be converted to the String '&{catty 10}'
	crd := eval.Wrap(c, &resource.CrdResource{
		Name: "catty",
		Age:  10,
	})
	ik := invokable(t)
	require.NotPanics(t, func() { ik.Invoke("Foo::Foo3", "create", crd) })
}

func TestRegisterServer_WithHandlerRegistration(t *testing.T) {
	// the bad/main.go attempts to register a handler in one of two different ways
	// but fails
	cmd := exec.Command("/usr/local/bin/go", "run", "../cmd/server/bad/main.go")
	_, err := Load(cmd)
	require.NoError(t, err)
}

func TestRegisterServer_TwoReturnValues(t *testing.T) {
	// the bad2/main.go attempts to register an api which returns a tuple (string, string)
	// but fails
	cmd := exec.Command("/usr/local/bin/go", "run", "../cmd/server/bad2/main.go")
	_, err := Load(cmd)
	require.NoError(t, err)
}
func invokable(t *testing.T) serviceapi.Invokable {
	cmd := exec.Command("/usr/local/bin/go", "run", "../cmd/server/main.go")
	server, err := Load(cmd)
	require.NoError(t, err)
	return server
}
