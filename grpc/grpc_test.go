package grpc

import (
	"fmt"
	"os/exec"
	"reflect"
	"testing"

	eval "github.com/puppetlabs/go-evaluator/eval"
	types "github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-servicesdk/cmd/server/resource"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/stretchr/testify/require"

	// Ensure initialization of pcore
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

func TestInvoke(t *testing.T) {
	eval.Puppet.Do(func(c eval.Context) {
		ik := invokable(c, t)
		actual := ik.Invoke("Foo::Foo", "hello", types.WrapString("Cassie"))
		require.Equal(t, "Hello Cassie!", actual.String())
	})
}

func TestInvoke_CanReturnStructType(t *testing.T) {
	eval.Puppet.Do(func(c eval.Context) {
		// This read should return a structured actual state but returns a string
		ik := invokable(c, t)
		actual := ik.Invoke("Foo::CrdHandler", "read", types.WrapString("1234"))
		require.NotEqual(t, "*types.StringValue", fmt.Sprintf("%T", actual))
		fmt.Println(eval.ToPrettyString(actual))
	})
}

func TestInvoke_FuncReturnsOneArgError(t *testing.T) {
	eval.Puppet.Do(func(c eval.Context) {
		// "Delete" returns only an error but CallGoReflected expects either:
		// - a single non-error return type, or
		// - a single non-error return type and a single error
		//
		// This call returns no error but causes an index out of bounds
		ik := invokable(c, t)
		require.NotPanics(t, func() { ik.Invoke("Foo::CrdHandler", "delete", types.WrapString("1234")) })
	})
}

func TestInvoke_FuncReturnsStringAndError(t *testing.T) {
	eval.Puppet.Do(func(c eval.Context) {
		// Similar to the example above
		// "Delete2" returns a single non-error return type and a single error
		// which meets the one arg, one error limitation
		//
		// When the error is nil the following occurs: panic: interface conversion: interface is nil, not error
		ik := invokable(c, t)
		var actual eval.Value
		require.NotPanics(t, func() { actual = ik.Invoke("Foo::CrdHandler", "delete2", types.WrapString("1234")) })
		fmt.Println(eval.ToPrettyString(actual))
	})
}

func TestInvoke_FuncReturnsStringAndError_ReturnsActualError(t *testing.T) {
	eval.Puppet.Do(func(c eval.Context) {
		// Similar to the example above
		// "Delete3" returns a single non-error return type and a single error
		// which meets the one arg, one error limitation
		//
		// When the error is *not* nil it is not treated as an application error,
		// rather the invocation is deemed to have failed. We can't differentiate
		// between a genesis error and a user-application error. We don't propagate
		// the errors through to the client-side caller.
		ik := invokable(c, t)
		var actual eval.Value
		require.NotPanics(t, func() { actual = ik.Invoke("Foo::CrdHandler", "delete3", types.WrapString("1234")) })
		fmt.Println(eval.ToPrettyString(actual))
	})
}

func TestInvoke_CanReceiveStruct(t *testing.T) {
	eval.Puppet.Do(func(c eval.Context) {
		// It would be great to be able to create the eval.Value needed to call create
		// This example serializes the structure to a string and reports this warning:
		// [0] contains a Runtime value. It will be converted to the String '&{catty 10}'
		ik := invokable(c, t)

		// TH: In order for the Wrap to know eval.Type the resource.CrdResource is, this must be
		// registered with the ImplementationRegistry prior to the Wrap call. The RegisterAPI
		// function does this automatically in the Plugin. Here we need to do it manually.

		// Since the Context is aware of the type (it's loaded from the Plugin using its
		// Metadata() function), a parse of the type name will now return the actual type.
		crdResourcType := c.ParseType2(`Foo::CrdResource`)

		// TH: Now register the association between the eval.Type and the Go type
		c.ImplementationRegistry().RegisterType(c, crdResourcType, reflect.TypeOf(&resource.CrdResource{}))

		// TH: The crd will now be a proper PuppetObject value instead of a Runtime value. The
		// latter is always converted to a String. Hence the earlier error
		crd := eval.Wrap(c, &resource.CrdResource{
			Name: "catty",
			Age:  10,
		})
		require.NotPanics(t, func() { ik.Invoke("Foo::CrdHandler", "create", crd) })
	})
}

func TestRegisterServer_WithHandlerRegistration(t *testing.T) {
	eval.Puppet.Do(func(c eval.Context) {
		// the bad/main.go attempts to register a handler in one of two different ways
		// but fails
		cmd := exec.Command("go", "run", "../cmd/server/bad/main.go")
		_, err := Load(cmd)
		require.NoError(t, err)
	})
}

func TestRegisterServer_TwoReturnValues(t *testing.T) {
	eval.Puppet.Do(func(c eval.Context) {
		ik := invokable(c, t)
		actual := ik.Invoke("Foo::Bar", "hello", types.WrapString("Tibbs"))
		fmt.Println(eval.ToPrettyString(actual))
		actualType := fmt.Sprintf("%T", actual)
		require.NotEqual(t, "*types.errorObj", actualType)
	})
}

func invokable(c eval.Context, t *testing.T) serviceapi.Service {
	cmd := exec.Command("go", "run", "../cmd/server/main.go")
	server, err := Load(cmd)
	require.NoError(t, err)

	// TH: Ensure that the eval.Context is aware of all types exported by the Plug-in
	ts, _ := server.Metadata()
	c.AddTypes(ts)
	return server
}
