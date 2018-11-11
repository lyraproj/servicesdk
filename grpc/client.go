package grpc

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/proto"
	"github.com/puppetlabs/go-evaluator/serialization"
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
	args := serialization.NewToDataConverter(eval.EMPTY_MAP).Convert(types.WrapValues(arguments))
	rq := servicepb.InvokeRequest{
		Identifier: identifier,
		Method: name,
		Arguments: proto.ToPBData(args),
	}
	rr, err := c.client.Invoke(c.ctx, &rq)
	if err != nil {
		panic(err)
	}
	return serialization.NewFromDataConverter(c.ctx, eval.EMPTY_MAP).Convert(proto.FromPBData(rr))
}

func (c *Client) Metadata() (typeSet eval.TypeSet, definitions []serviceapi.Definition) {
	rq := servicepb.MetadataRequest{}
	rr, err := c.client.Metadata(c.ctx, &rq)
	if err != nil {
		panic(err)
	}
	fdc := serialization.NewFromDataConverter(c.ctx, eval.EMPTY_MAP)
	typeSet = fdc.Convert(proto.FromPBData(rr.GetTypeset())).(eval.TypeSet)

	ds := fdc.Convert(proto.FromPBData(rr.GetDefinitions())).(eval.List)
	definitions = make([]serviceapi.Definition, ds.Len())
	ds.EachWithIndex(func(d eval.Value, i int) { definitions[i] = d.(serviceapi.Definition) })
	return
}
