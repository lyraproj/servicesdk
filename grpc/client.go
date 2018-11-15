package grpc

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/puppetlabs/go-servicesdk/servicepb"
	"google.golang.org/grpc"
)

type Plugin struct {
}

func (a *Plugin) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	return fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, clientConn *grpc.ClientConn) (interface{}, error) {
	return &Client{ctx: ctx.(eval.Context), client: servicepb.NewDefinitionServiceClient(clientConn) }, nil
}

type Client struct {
	ctx eval.Context
	client servicepb.DefinitionServiceClient
}

func (c *Client) Invoke(identifier, name string, arguments ...eval.Value) eval.Value {
	rq := servicepb.InvokeRequest{
		Identifier: identifier,
		Method: name,
		Arguments: ToDataPB(types.WrapValues(arguments)),
	}
	rr, err := c.client.Invoke(c.ctx, &rq)
	if err != nil {
		panic(err)
	}
	return FromDataPB(c.ctx, rr)
}

func (c *Client) Metadata() (typeSet eval.TypeSet, definitions []serviceapi.Definition) {
	rq := servicepb.MetadataRequest{}
	rr, err := c.client.Metadata(c.ctx, &rq)
	if err != nil {
		panic(err)
	}
	typeSet = FromDataPB(c.ctx, rr.GetTypeset()).(eval.TypeSet)
	ds := FromDataPB(c.ctx, rr.GetDefinitions()).(eval.List)
	definitions = make([]serviceapi.Definition, ds.Len())
	ds.EachWithIndex(func(d eval.Value, i int) { definitions[i] = d.(serviceapi.Definition) })
	return
}

func (c *Client) State(identifier string, input eval.OrderedMap) eval.PuppetObject {
	rq := servicepb.StateRequest{Identifier: identifier, Input: ToDataPB(input)}
	rr, err := c.client.State(c.ctx, &rq)
	if err != nil {
		panic(err)
	}
	return FromDataPB(c.ctx, rr).(eval.PuppetObject)
}