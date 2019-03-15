package grpc

import (
	"fmt"
	"net/rpc"

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
)

type Server struct {
	ctx  px.Context
	impl serviceapi.Service
}

func (s *Server) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, s)
}

func (s *Server) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no RPC client implementation for rpc`, s)
}

func (s *Server) GRPCServer(broker *plugin.GRPCBroker, impl *grpc.Server) error {
	servicepb.RegisterDefinitionServiceServer(impl, s)
	return nil
}

func (s *Server) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, s)
}

func (s *Server) Do(doer func(c px.Context)) (err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(issue.Reported); ok {
				err = e
			} else {
				panic(x)
			}
		}
	}()
	c := s.ctx.Fork()
	threadlocal.Init()
	threadlocal.Set(px.PuppetContextKey, c)
	doer(c)
	return nil
}

func (s *Server) Identity(context.Context, *servicepb.EmptyRequest) (result *datapb.Data, err error) {
	err = s.Do(func(c px.Context) {
		result = ToDataPB(s.impl.Identifier(c))
	})
	return
}

func (s *Server) Invoke(_ context.Context, r *servicepb.InvokeRequest) (result *datapb.Data, err error) {
	err = s.Do(func(c px.Context) {
		wrappedArgs := FromDataPB(c, r.Arguments)
		arguments := wrappedArgs.(*types.Array).AppendTo([]px.Value{})
		rrr := s.impl.Invoke(
			c,
			r.Identifier,
			r.Method,
			arguments...)
		result = ToDataPB(rrr)
	})
	return
}

func (s *Server) Metadata(_ context.Context, r *servicepb.EmptyRequest) (result *servicepb.MetadataResponse, err error) {
	err = s.Do(func(c px.Context) {
		ts, ds := s.impl.Metadata(c)
		vs := make([]px.Value, len(ds))
		for i, d := range ds {
			vs[i] = d
		}
		result = &servicepb.MetadataResponse{Typeset: ToDataPB(ts), Definitions: ToDataPB(types.WrapValues(vs))}
	})
	return
}

func (s *Server) State(_ context.Context, r *servicepb.StateRequest) (result *datapb.Data, err error) {
	err = s.Do(func(c px.Context) {
		result = ToDataPB(s.impl.State(c, r.Identifier, FromDataPB(c, r.Input).(px.OrderedMap)))
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
	logger := hclog.Default()
	cfg := &plugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]plugin.Plugin{
			"server": &Server{ctx: c, impl: s},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	}
	name := s.Identifier(c).Name()
	logger.Debug("Starting to serve", "name", name)
	plugin.Serve(cfg)
	logger.Debug("Done serving", "name", name)
}
