// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
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
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

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
	ProcessTask(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[Task, emptypb.Empty], error)
}

type clusterClient struct {
	cc grpc.ClientConnInterface
}

func NewClusterClient(cc grpc.ClientConnInterface) ClusterClient {
	return &clusterClient{cc}
}

func (c *clusterClient) Join(ctx context.Context, in *NodeProperties, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Cluster_Join_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterClient) UpdateNode(ctx context.Context, in *NodeProperties, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Cluster_UpdateNode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterClient) Echo(ctx context.Context, in *Ping, opts ...grpc.CallOption) (*Pong, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Pong)
	err := c.cc.Invoke(ctx, Cluster_Echo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterClient) ProcessTask(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[Task, emptypb.Empty], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Cluster_ServiceDesc.Streams[0], Cluster_ProcessTask_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[Task, emptypb.Empty]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Cluster_ProcessTaskClient = grpc.ClientStreamingClient[Task, emptypb.Empty]

// ClusterServer is the server API for Cluster service.
// All implementations must embed UnimplementedClusterServer
// for forward compatibility.
type ClusterServer interface {
	Join(context.Context, *NodeProperties) (*emptypb.Empty, error)
	UpdateNode(context.Context, *NodeProperties) (*emptypb.Empty, error)
	Echo(context.Context, *Ping) (*Pong, error)
	ProcessTask(grpc.ClientStreamingServer[Task, emptypb.Empty]) error
	mustEmbedUnimplementedClusterServer()
}

// UnimplementedClusterServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedClusterServer struct{}

func (UnimplementedClusterServer) Join(context.Context, *NodeProperties) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Join not implemented")
}
func (UnimplementedClusterServer) UpdateNode(context.Context, *NodeProperties) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateNode not implemented")
}
func (UnimplementedClusterServer) Echo(context.Context, *Ping) (*Pong, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Echo not implemented")
}
func (UnimplementedClusterServer) ProcessTask(grpc.ClientStreamingServer[Task, emptypb.Empty]) error {
	return status.Errorf(codes.Unimplemented, "method ProcessTask not implemented")
}
func (UnimplementedClusterServer) mustEmbedUnimplementedClusterServer() {}
func (UnimplementedClusterServer) testEmbeddedByValue()                 {}

// UnsafeClusterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClusterServer will
// result in compilation errors.
type UnsafeClusterServer interface {
	mustEmbedUnimplementedClusterServer()
}

func RegisterClusterServer(s grpc.ServiceRegistrar, srv ClusterServer) {
	// If the following call pancis, it indicates UnimplementedClusterServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
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
	return srv.(ClusterServer).ProcessTask(&grpc.GenericServerStream[Task, emptypb.Empty]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Cluster_ProcessTaskServer = grpc.ClientStreamingServer[Task, emptypb.Empty]

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
