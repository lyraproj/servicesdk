package grpc

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/servicepb"
	"google.golang.org/grpc"
	"os/exec"
)

var handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "7468697320697320616e20616d617a696e67206d6167696320636f6f6b69652c206e6f6d206e6f6d206e6f6d",
}

type PluginClient struct {
	plugin.NetRPCUnsupportedPlugin
}

func (a *PluginClient) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	return fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *PluginClient) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, clientConn *grpc.ClientConn) (interface{}, error) {
	return &Client{client: servicepb.NewDefinitionServiceClient(clientConn)}, nil
}

type Client struct {
	client servicepb.DefinitionServiceClient
}

func (c *Client) Identifier(ctx eval.Context) eval.TypedName {
	rr, err := c.client.Identity(ctx, &servicepb.EmptyRequest{})
	if err != nil {
		panic(err)
	}
	return FromDataPB(ctx, rr).(eval.TypedName)
}

func (c *Client) Invoke(ctx eval.Context, identifier, name string, arguments ...eval.Value) eval.Value {
	rq := servicepb.InvokeRequest{
		Identifier: identifier,
		Method:     name,
		Arguments:  ToDataPB(types.WrapValues(arguments)),
	}
	rr, err := c.client.Invoke(ctx, &rq)
	if err != nil {
		panic(err)
	}
	result := FromDataPB(ctx, rr)
	if eo, ok := result.(eval.ErrorObject); ok {
		panic(eval.Error(WF_INVOCATION_ERROR, issue.H{`identifier`: identifier, `name`: name, `code`: eo.IssueCode(), `message`: eo.Message()}))
	}
	return result
}

func (c *Client) Metadata(ctx eval.Context) (typeSet eval.TypeSet, definitions []serviceapi.Definition) {
	rr, err := c.client.Metadata(ctx, &servicepb.EmptyRequest{})
	if err != nil {
		panic(err)
	}
	typeSet = FromDataPB(ctx, rr.GetTypeset()).(eval.TypeSet)
	ds := FromDataPB(ctx, rr.GetDefinitions()).(eval.List)
	definitions = make([]serviceapi.Definition, ds.Len())
	ds.EachWithIndex(func(d eval.Value, i int) { definitions[i] = d.(serviceapi.Definition) })
	return
}

func (c *Client) State(ctx eval.Context, identifier string, input eval.OrderedMap) eval.PuppetObject {
	rq := servicepb.StateRequest{Identifier: identifier, Input: ToDataPB(input)}
	rr, err := c.client.State(ctx, &rq)
	if err != nil {
		panic(err)
	}
	return FromDataPB(ctx, rr).(eval.PuppetObject)
}

// Load  ...
func Load(cmd *exec.Cmd, logger hclog.Logger) (serviceapi.Service, error) {
	if logger == nil {
		logger = hclog.Default()
	}
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]plugin.Plugin{
			"server": &PluginClient{},
		},
		Managed:          true,
		Cmd:              cmd,
		Logger:           logger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	grpcClient, err := client.Client()
	if err != nil {
		hclog.Default().Error("error creating GRPC client", "error", err)
		return nil, err
	}

	// Request the plugin
	pluginName := "server"
	raw, err := grpcClient.Dispense(pluginName)
	if err != nil {
		hclog.Default().Error("error dispensing plugin", "plugin", pluginName, "error", err)
		return nil, err
	}
	return raw.(serviceapi.Service), nil
}
