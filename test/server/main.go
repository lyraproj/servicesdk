package main

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/lyraproj/servicesdk/test/shared"
	"google.golang.org/grpc"
	"os"
)

type Server struct {
}

type MyPlugin struct {
	plugin.NetRPCUnsupportedPlugin
}

func (MyPlugin) GRPCServer(b *plugin.GRPCBroker, s *grpc.Server) error {
	shared.RegisterHelloServiceServer(s, &Server{})
	return nil
}

func (MyPlugin) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (interface{}, error) {
	panic("no client")
}

func (Server) Hello(_ context.Context, m *shared.HelloMsg) (*shared.HelloMsg, error) {
	return &shared.HelloMsg{Hello: "Hello " + m.Hello}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         map[string]plugin.Plugin{"hello": &MyPlugin{}},
		GRPCServer:      plugin.DefaultGRPCServer,
	})
	os.Exit(0)
}
