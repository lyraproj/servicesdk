package grpc

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/proto"
	"github.com/puppetlabs/go-evaluator/serialization"
	"github.com/puppetlabs/go-evaluator/threadlocal"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/puppetlabs/go-servicesdk/servicepb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net/rpc"

	// Ensure that pcore is initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
	"github.com/puppetlabs/go-evaluator/types"
)

type GRPCServer struct {
	ctx  eval.Context
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

func (a *GRPCServer) Do(doer func(c eval.Context)) (err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(issue.Reported); ok {
				err = e
			} else {
				panic(x)
			}
		}
	}()
	threadlocal.Init()
	threadlocal.Set(eval.PuppetContextKey, a.ctx)
	doer(a.ctx)
	return nil
}

func (d *GRPCServer) Invoke(_ context.Context, r *servicepb.InvokeRequest) (result *datapb.Data, err error) {
	err = d.Do(func(c eval.Context) {
		wrappedArgs := FromDataPB(c, r.Arguments)
		arguments := wrappedArgs.(*types.ArrayValue).AppendTo([]eval.Value{})
		rrr := d.impl.Invoke(
			r.Identifier,
			r.Method,
			arguments...)
		result = ToDataPB(rrr)
	})
	return
}

func (d *GRPCServer) Metadata(_ context.Context, r *servicepb.MetadataRequest) (result *servicepb.MetadataResponse, err error) {
	err = d.Do(func(c eval.Context) {
		ts, ds := d.impl.Metadata()
		vs := make([]eval.Value, len(ds))
		for i, d := range ds {
			vs[i] = d
		}
		result = &servicepb.MetadataResponse{Typeset: ToDataPB(ts), Definitions: ToDataPB(types.WrapValues(vs))}
	})
	return
}

func (d *GRPCServer) State(_ context.Context, r *servicepb.StateRequest) (result *datapb.Data, err error) {
	err = d.Do(func(c eval.Context) {
		result = ToDataPB(d.impl.State(r.Identifier, FromDataPB(c, r.Input).(eval.OrderedMap)))
	})
	return
}

func ToDataPB(v eval.Value) *datapb.Data {
	return proto.ToPBData(serialization.NewToDataConverter(eval.EMPTY_MAP).Convert(v))
}

func FromDataPB(c eval.Context, d *datapb.Data) eval.Value {
	return serialization.NewFromDataConverter(c, eval.EMPTY_MAP).Convert(proto.FromPBData(d))
}

// Serve the supplied Server as a go-plugin
func Serve(c eval.Context, s serviceapi.Service) {
	cfg := &plugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]plugin.Plugin{
			"server": &GRPCServer{ctx: c, impl: s},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     hclog.Default(),
	}
	plugin.Serve(cfg)
}
