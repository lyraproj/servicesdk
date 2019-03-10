package grpc

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/lyraproj/data-protobuf/datapb"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/proto"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/serialization"
	"github.com/lyraproj/pcore/threadlocal"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/servicepb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net/rpc"

	// Ensure that pcore is initialized
	_ "github.com/lyraproj/pcore/pcore"
)

type GRPCServer struct {
	ctx  px.Context
	impl serviceapi.Service
}

func (a *GRPCServer) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *GRPCServer) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no RPC client implementation for rpc`, a)
}

func (a *GRPCServer) GRPCServer(broker *plugin.GRPCBroker, impl *grpc.Server) error {
	servicepb.RegisterDefinitionServiceServer(impl, a)
	return nil
}

func (a *GRPCServer) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, a)
}

func (a *GRPCServer) Do(doer func(c px.Context)) (err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(issue.Reported); ok {
				err = e
			} else {
				panic(x)
			}
		}
	}()
	c := a.ctx.Fork()
	threadlocal.Init()
	threadlocal.Set(px.PuppetContextKey, c)
	doer(c)
	return nil
}

func (d *GRPCServer) Identity(context.Context, *servicepb.EmptyRequest) (result *datapb.Data, err error) {
	err = d.Do(func(c px.Context) {
		result = ToDataPB(d.impl.Identifier(c))
	})
	return
}

func (d *GRPCServer) Invoke(_ context.Context, r *servicepb.InvokeRequest) (result *datapb.Data, err error) {
	err = d.Do(func(c px.Context) {
		wrappedArgs := FromDataPB(c, r.Arguments)
		arguments := wrappedArgs.(*types.Array).AppendTo([]px.Value{})
		rrr := d.impl.Invoke(
			c,
			r.Identifier,
			r.Method,
			arguments...)
		result = ToDataPB(rrr)
	})
	return
}

func (d *GRPCServer) Metadata(_ context.Context, r *servicepb.EmptyRequest) (result *servicepb.MetadataResponse, err error) {
	err = d.Do(func(c px.Context) {
		ts, ds := d.impl.Metadata(c)
		vs := make([]px.Value, len(ds))
		for i, d := range ds {
			vs[i] = d
		}
		result = &servicepb.MetadataResponse{Typeset: ToDataPB(ts), Definitions: ToDataPB(types.WrapValues(vs))}
	})
	return
}

func (d *GRPCServer) State(_ context.Context, r *servicepb.StateRequest) (result *datapb.Data, err error) {
	err = d.Do(func(c px.Context) {
		result = ToDataPB(d.impl.State(c, r.Identifier, FromDataPB(c, r.Input).(px.OrderedMap)))
	})
	return
}

func ToDataPB(v px.Value) *datapb.Data {
	if v == nil {
		return nil
	}
	pc := proto.NewProtoConsumer()
	serialization.NewSerializer(pcore.RootContext(), px.EmptyMap).Convert(v, pc)
	return pc.Value()
}

func FromDataPB(c px.Context, d *datapb.Data) px.Value {
	if d == nil {
		return nil
	}
	ds := serialization.NewDeserializer(c, px.EmptyMap)
	proto.ConsumePBData(d, ds)
	return ds.Value()
}

// Serve the supplied Server as a go-plugin
func Serve(c px.Context, s serviceapi.Service) {
	cfg := &plugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]plugin.Plugin{
			"server": &GRPCServer{ctx: c, impl: s},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     hclog.Default(),
	}
	id := s.Identifier(c)
	log.Printf("Starting to serve %s\n", id)
	plugin.Serve(cfg)
	log.Printf("Done serve %s\n", id)
}
