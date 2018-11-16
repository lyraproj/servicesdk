package grpc

import (
	"context"
	"fmt"
	"net/rpc"
	"os/exec"

	//FIXME same import twice
	goplugin "github.com/hashicorp/go-plugin"
	plugin "github.com/hashicorp/go-plugin"

	hclog "github.com/hashicorp/go-hclog"

	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-servicesdk/service"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/puppetlabs/go-servicesdk/servicepb"

	//FIXME same import twice
	"google.golang.org/grpc"
	gogrpc "google.golang.org/grpc"
)

var handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "7468697320697320616e20616d617a696e67206d6167696320636f6f6b69652c206e6f6d206e6f6d206e6f6d",
}

type Plugin struct {
}

func (a *Plugin) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	return fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, clientConn *grpc.ClientConn) (interface{}, error) {
	return &Client{ctx: ctx.(eval.Context), client: servicepb.NewDefinitionServiceClient(clientConn)}, nil
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

// Serve the supplied Server as a go-plugin
func Serve(s *service.Server) {
	cfg := &goplugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]goplugin.Plugin{
			"server": &goPlugin{Impl: s},
		},
		GRPCServer: goplugin.DefaultGRPCServer,
		Logger:     hclog.Default(),
	}
	goplugin.Serve(cfg)
}

// Load  ...
func Load(cmd *exec.Cmd) (serviceapi.Invokable, error) {

	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]goplugin.Plugin{
			"server": &goPlugin{},
		},
		Cmd:              cmd,
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
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
	invokable := raw.(serviceapi.Invokable)
	return invokable, nil
}

type goPlugin struct {
	Impl *service.Server
}

// Server returns a Provider Resource RPC server. (Not supported)
func (*goPlugin) Server(*goplugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf("Plugin does not support net/rpc server")
}

// Client returns a Provider Resource RPC client. (Not supported)
func (*goPlugin) Client(b *goplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf("Plugin does not support net/rpc client")
}

func (p *goPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *gogrpc.Server) error {
	servicepb.RegisterDefinitionServiceServer(s, &Server{impl: p.Impl})
	return nil
}

func (*goPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, clientConn *gogrpc.ClientConn) (interface{}, error) {
	return &Client{
		ctx:    eval.Puppet.RootContext(), // FIXME should be the eval context with the types registered
		client: servicepb.NewDefinitionServiceClient(clientConn),
	}, nil
}
