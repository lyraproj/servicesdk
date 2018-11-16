package grpc

import (
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/proto"
	"github.com/puppetlabs/go-evaluator/serialization"
	"github.com/puppetlabs/go-servicesdk/service"
	"github.com/puppetlabs/go-servicesdk/servicepb"
	"golang.org/x/net/context"

	// Ensure that pcore is initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
	"github.com/puppetlabs/go-evaluator/types"
)

type Server struct {
	impl *service.Server
}

func NewServer(impl *service.Server) *Server {
	return &Server{impl: impl}
}

func (d *Server) Invoke(c context.Context, r *servicepb.InvokeRequest) (result *datapb.Data, err error) {
	err = eval.Puppet.TryWithParent(c, func(ec eval.Context) error {
		result = ToDataPB(d.impl.Invoke(
			r.Identifier,
			r.Method,
			FromDataPB(ec, r.Arguments).(eval.OrderedMap)))
		return nil
	})
	return result, err
}

func (d *Server) Metadata(c context.Context, r *servicepb.MetadataRequest) (result *servicepb.MetadataResponse, err error) {
	err = eval.Puppet.TryWithParent(c, func(ec eval.Context) error {
		ts, ds := d.impl.Metadata()
		vs := make([]eval.Value, len(ds))
		for i, d := range ds {
			vs[i] = d
		}
		result = &servicepb.MetadataResponse{Typeset: ToDataPB(ts), Definitions: ToDataPB(types.WrapValues(vs))}
		return nil
	})
	return result, err
}

func (d *Server) State(c context.Context, r *servicepb.StateRequest) (result *datapb.Data, err error) {
	err = eval.Puppet.TryWithParent(c, func(ec eval.Context) error {
		result = ToDataPB(d.impl.State(
			r.Identifier,
			FromDataPB(ec, r.Input).(eval.OrderedMap)))
		return nil
	})
	return result, err
}

func ToDataPB(v eval.Value) *datapb.Data {
	return proto.ToPBData(serialization.NewToDataConverter(eval.EMPTY_MAP).Convert(v))
}

func FromDataPB(c eval.Context, d *datapb.Data) eval.Value {
	return serialization.NewFromDataConverter(c, eval.EMPTY_MAP).Convert(proto.FromPBData(d))
}
