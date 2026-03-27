package echo

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	ServiceName    = "example.echo.EchoService"
	EchoMethodName = "Echo"
	EchoFullMethod = "/" + ServiceName + "/" + EchoMethodName
)

type ServiceServer interface {
	Echo(context.Context, *wrapperspb.StringValue) (*wrapperspb.StringValue, error)
}

func RegisterServiceServer(registrar grpc.ServiceRegistrar, srv ServiceServer) {
	registrar.RegisterService(&grpc.ServiceDesc{
		ServiceName: ServiceName,
		HandlerType: (*ServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: EchoMethodName,
				Handler:    echoHandler,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "manual-echo",
	}, srv)
}

func echoHandler(
	srv interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	in := new(wrapperspb.StringValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).Echo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EchoFullMethod,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).Echo(ctx, req.(*wrapperspb.StringValue))
	}
	return interceptor(ctx, in, info, handler)
}

func CallEcho(ctx context.Context, conn grpc.ClientConnInterface, msg string) (string, error) {
	req := wrapperspb.String(msg)
	resp := new(wrapperspb.StringValue)
	if err := conn.Invoke(ctx, EchoFullMethod, req, resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}
