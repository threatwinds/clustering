// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: clustering.proto

package clustering

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Cluster_Join_FullMethodName        = "/clustering.Cluster/Join"
	Cluster_UpdateNode_FullMethodName  = "/clustering.Cluster/UpdateNode"
	Cluster_Echo_FullMethodName        = "/clustering.Cluster/Echo"
	Cluster_ProcessTask_FullMethodName = "/clustering.Cluster/ProcessTask"
)

// ClusterClient is the client API for Cluster service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClusterClient interface {
	Join(ctx context.Context, in *NodeProperties, opts ...grpc.CallOption) (*emptypb.Empty, error)
	UpdateNode(ctx context.Context, in *NodeProperties, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Echo(ctx context.Context, in *Ping, opts ...grpc.CallOption) (*Pong, error)
	ProcessTask(ctx context.Context, opts ...grpc.CallOption) (Cluster_ProcessTaskClient, error)
}

type clusterClient struct {
	cc grpc.ClientConnInterface
}

func NewClusterClient(cc grpc.ClientConnInterface) ClusterClient {
	return &clusterClient{cc}
}

func (c *clusterClient) Join(ctx context.Context, in *NodeProperties, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Cluster_Join_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterClient) UpdateNode(ctx context.Context, in *NodeProperties, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Cluster_UpdateNode_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterClient) Echo(ctx context.Context, in *Ping, opts ...grpc.CallOption) (*Pong, error) {
	out := new(Pong)
	err := c.cc.Invoke(ctx, Cluster_Echo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterClient) ProcessTask(ctx context.Context, opts ...grpc.CallOption) (Cluster_ProcessTaskClient, error) {
	stream, err := c.cc.NewStream(ctx, &Cluster_ServiceDesc.Streams[0], Cluster_ProcessTask_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &clusterProcessTaskClient{stream}
	return x, nil
}

type Cluster_ProcessTaskClient interface {
	Send(*Task) error
	CloseAndRecv() (*emptypb.Empty, error)
	grpc.ClientStream
}

type clusterProcessTaskClient struct {
	grpc.ClientStream
}

func (x *clusterProcessTaskClient) Send(m *Task) error {
	return x.ClientStream.SendMsg(m)
}

func (x *clusterProcessTaskClient) CloseAndRecv() (*emptypb.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(emptypb.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ClusterServer is the server API for Cluster service.
// All implementations must embed UnimplementedClusterServer
// for forward compatibility
type ClusterServer interface {
	Join(context.Context, *NodeProperties) (*emptypb.Empty, error)
	UpdateNode(context.Context, *NodeProperties) (*emptypb.Empty, error)
	Echo(context.Context, *Ping) (*Pong, error)
	ProcessTask(Cluster_ProcessTaskServer) error
	mustEmbedUnimplementedClusterServer()
}

// UnimplementedClusterServer must be embedded to have forward compatible implementations.
type UnimplementedClusterServer struct {
}

func (UnimplementedClusterServer) Join(context.Context, *NodeProperties) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Join not implemented")
}
func (UnimplementedClusterServer) UpdateNode(context.Context, *NodeProperties) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateNode not implemented")
}
func (UnimplementedClusterServer) Echo(context.Context, *Ping) (*Pong, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Echo not implemented")
}
func (UnimplementedClusterServer) ProcessTask(Cluster_ProcessTaskServer) error {
	return status.Errorf(codes.Unimplemented, "method ProcessTask not implemented")
}
func (UnimplementedClusterServer) mustEmbedUnimplementedClusterServer() {}

// UnsafeClusterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClusterServer will
// result in compilation errors.
type UnsafeClusterServer interface {
	mustEmbedUnimplementedClusterServer()
}

func RegisterClusterServer(s grpc.ServiceRegistrar, srv ClusterServer) {
	s.RegisterService(&Cluster_ServiceDesc, srv)
}

func _Cluster_Join_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeProperties)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServer).Join(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Cluster_Join_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServer).Join(ctx, req.(*NodeProperties))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cluster_UpdateNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeProperties)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServer).UpdateNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Cluster_UpdateNode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServer).UpdateNode(ctx, req.(*NodeProperties))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cluster_Echo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Ping)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServer).Echo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Cluster_Echo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServer).Echo(ctx, req.(*Ping))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cluster_ProcessTask_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ClusterServer).ProcessTask(&clusterProcessTaskServer{stream})
}

type Cluster_ProcessTaskServer interface {
	SendAndClose(*emptypb.Empty) error
	Recv() (*Task, error)
	grpc.ServerStream
}

type clusterProcessTaskServer struct {
	grpc.ServerStream
}

func (x *clusterProcessTaskServer) SendAndClose(m *emptypb.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *clusterProcessTaskServer) Recv() (*Task, error) {
	m := new(Task)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Cluster_ServiceDesc is the grpc.ServiceDesc for Cluster service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Cluster_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "clustering.Cluster",
	HandlerType: (*ClusterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Join",
			Handler:    _Cluster_Join_Handler,
		},
		{
			MethodName: "UpdateNode",
			Handler:    _Cluster_UpdateNode_Handler,
		},
		{
			MethodName: "Echo",
			Handler:    _Cluster_Echo_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ProcessTask",
			Handler:       _Cluster_ProcessTask_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "clustering.proto",
}
