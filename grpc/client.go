package grpc

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/puppetlabs/go-servicesdk/servicepb"
	"google.golang.org/grpc"
	"net/rpc"
	"os/exec"
	"os"
)

var handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "7468697320697320616e20616d617a696e67206d6167696320636f6f6b69652c206e6f6d206e6f6d206e6f6d",
}

type PluginClient struct {
}

func (a *PluginClient) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *PluginClient) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no RPC client implementation for rpc`, a)
}

func (a *PluginClient) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	return fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *PluginClient) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, clientConn *grpc.ClientConn) (interface{}, error) {
	return &Client{ctx: eval.CurrentContext(), client: servicepb.NewDefinitionServiceClient(clientConn)}, nil
}

type Client struct {
	ctx    eval.Context
	client servicepb.DefinitionServiceClient
}

func (c *Client) Invoke(identifier, name string, arguments ...eval.Value) eval.Value {
	rq := servicepb.InvokeRequest{
		Identifier: identifier,
		Method:     name,
		Arguments:  ToDataPB(types.WrapValues(arguments)),
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

// Load  ...
func Load(cmd *exec.Cmd) (serviceapi.Service, error) {

	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Debug,
		Output:     os.Stdout,
		JSONFormat: true,
	})
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]plugin.Plugin{
			"server": &PluginClient{},
		},
		Cmd:              cmd,
		Logger: logger,
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
