package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/lyraproj/servicesdk/test/shared"
	"google.golang.org/grpc"
	"os"
	"os/exec"
)

type Client struct {
	client shared.HelloServiceClient
}

func (c *Client) Hello(msg string) string {
	r, err := c.client.Hello(context.Background(), &shared.HelloMsg{Hello: "world"})
	if err != nil {
		panic(err)
	}
	return r.Hello
}

type MyPlugin struct {
	plugin.NetRPCUnsupportedPlugin
}

func (MyPlugin) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	panic("No server")
}

func (MyPlugin) GRPCClient(c context.Context, broker *plugin.GRPCBroker, clientConn *grpc.ClientConn) (interface{}, error) {
	return &Client{client: shared.NewHelloServiceClient(clientConn)}, nil
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println(wd)
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          map[string]plugin.Plugin{"hello": &MyPlugin{}},
		Cmd:              exec.Command("go", "run", "test/server/main.go", "--debug"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})
	defer client.Kill()

	cl, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := cl.Dispense("hello")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	hello := raw.(shared.HelloApp)
	fmt.Println(hello.Hello("world"))
	os.Exit(0)
}
