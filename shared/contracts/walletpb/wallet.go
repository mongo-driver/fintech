package walletpb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

const serviceName = "fintech.wallet.v1.WalletService"

type WalletServiceServer interface {
	CreateWallet(context.Context, *structpb.Struct) (*structpb.Struct, error)
	GetWallet(context.Context, *structpb.Struct) (*structpb.Struct, error)
	Deposit(context.Context, *structpb.Struct) (*structpb.Struct, error)
	Withdraw(context.Context, *structpb.Struct) (*structpb.Struct, error)
	Transfer(context.Context, *structpb.Struct) (*structpb.Struct, error)
	ListTransactions(context.Context, *structpb.Struct) (*structpb.Struct, error)
}

func RegisterWalletServiceServer(s grpc.ServiceRegistrar, srv WalletServiceServer) {
	s.RegisterService(&WalletService_ServiceDesc, srv)
}

type WalletServiceClient interface {
	CreateWallet(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error)
	GetWallet(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error)
	Deposit(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error)
	Withdraw(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error)
	Transfer(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error)
	ListTransactions(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error)
}

type walletServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWalletServiceClient(cc grpc.ClientConnInterface) WalletServiceClient {
	return &walletServiceClient{cc: cc}
}

func (c *walletServiceClient) CreateWallet(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	out := new(structpb.Struct)
	err := c.cc.Invoke(ctx, "/"+serviceName+"/CreateWallet", in, out, opts...)
	return out, err
}
func (c *walletServiceClient) GetWallet(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	out := new(structpb.Struct)
	err := c.cc.Invoke(ctx, "/"+serviceName+"/GetWallet", in, out, opts...)
	return out, err
}
func (c *walletServiceClient) Deposit(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	out := new(structpb.Struct)
	err := c.cc.Invoke(ctx, "/"+serviceName+"/Deposit", in, out, opts...)
	return out, err
}
func (c *walletServiceClient) Withdraw(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	out := new(structpb.Struct)
	err := c.cc.Invoke(ctx, "/"+serviceName+"/Withdraw", in, out, opts...)
	return out, err
}
func (c *walletServiceClient) Transfer(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	out := new(structpb.Struct)
	err := c.cc.Invoke(ctx, "/"+serviceName+"/Transfer", in, out, opts...)
	return out, err
}
func (c *walletServiceClient) ListTransactions(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	out := new(structpb.Struct)
	err := c.cc.Invoke(ctx, "/"+serviceName+"/ListTransactions", in, out, opts...)
	return out, err
}

func _WalletService_CreateWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(structpb.Struct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).CreateWallet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/CreateWallet"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).CreateWallet(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_GetWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(structpb.Struct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).GetWallet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/GetWallet"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).GetWallet(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_Deposit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(structpb.Struct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).Deposit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/Deposit"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).Deposit(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_Withdraw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(structpb.Struct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).Withdraw(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/Withdraw"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).Withdraw(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_Transfer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(structpb.Struct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).Transfer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/Transfer"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).Transfer(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_ListTransactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(structpb.Struct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).ListTransactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/ListTransactions"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).ListTransactions(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

var WalletService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: serviceName,
	HandlerType: (*WalletServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateWallet", Handler: _WalletService_CreateWallet_Handler},
		{MethodName: "GetWallet", Handler: _WalletService_GetWallet_Handler},
		{MethodName: "Deposit", Handler: _WalletService_Deposit_Handler},
		{MethodName: "Withdraw", Handler: _WalletService_Withdraw_Handler},
		{MethodName: "Transfer", Handler: _WalletService_Transfer_Handler},
		{MethodName: "ListTransactions", Handler: _WalletService_ListTransactions_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "wallet.proto",
}
