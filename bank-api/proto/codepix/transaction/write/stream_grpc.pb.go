// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.1
// source: proto/codepix/transaction/write/stream.proto

package write

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// StreamClient is the client API for Stream service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StreamClient interface {
	Start(ctx context.Context, opts ...grpc.CallOption) (Stream_StartClient, error)
	Confirm(ctx context.Context, opts ...grpc.CallOption) (Stream_ConfirmClient, error)
	Complete(ctx context.Context, opts ...grpc.CallOption) (Stream_CompleteClient, error)
	Fail(ctx context.Context, opts ...grpc.CallOption) (Stream_FailClient, error)
}

type streamClient struct {
	cc grpc.ClientConnInterface
}

func NewStreamClient(cc grpc.ClientConnInterface) StreamClient {
	return &streamClient{cc}
}

func (c *streamClient) Start(ctx context.Context, opts ...grpc.CallOption) (Stream_StartClient, error) {
	stream, err := c.cc.NewStream(ctx, &Stream_ServiceDesc.Streams[0], "/codepix.transaction.write.Stream/Start", opts...)
	if err != nil {
		return nil, err
	}
	x := &streamStartClient{stream}
	return x, nil
}

type Stream_StartClient interface {
	Send(*StartRequest) error
	Recv() (*StartReply, error)
	grpc.ClientStream
}

type streamStartClient struct {
	grpc.ClientStream
}

func (x *streamStartClient) Send(m *StartRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *streamStartClient) Recv() (*StartReply, error) {
	m := new(StartReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *streamClient) Confirm(ctx context.Context, opts ...grpc.CallOption) (Stream_ConfirmClient, error) {
	stream, err := c.cc.NewStream(ctx, &Stream_ServiceDesc.Streams[1], "/codepix.transaction.write.Stream/Confirm", opts...)
	if err != nil {
		return nil, err
	}
	x := &streamConfirmClient{stream}
	return x, nil
}

type Stream_ConfirmClient interface {
	Send(*ConfirmRequest) error
	Recv() (*ConfirmReply, error)
	grpc.ClientStream
}

type streamConfirmClient struct {
	grpc.ClientStream
}

func (x *streamConfirmClient) Send(m *ConfirmRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *streamConfirmClient) Recv() (*ConfirmReply, error) {
	m := new(ConfirmReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *streamClient) Complete(ctx context.Context, opts ...grpc.CallOption) (Stream_CompleteClient, error) {
	stream, err := c.cc.NewStream(ctx, &Stream_ServiceDesc.Streams[2], "/codepix.transaction.write.Stream/Complete", opts...)
	if err != nil {
		return nil, err
	}
	x := &streamCompleteClient{stream}
	return x, nil
}

type Stream_CompleteClient interface {
	Send(*CompleteRequest) error
	Recv() (*CompleteReply, error)
	grpc.ClientStream
}

type streamCompleteClient struct {
	grpc.ClientStream
}

func (x *streamCompleteClient) Send(m *CompleteRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *streamCompleteClient) Recv() (*CompleteReply, error) {
	m := new(CompleteReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *streamClient) Fail(ctx context.Context, opts ...grpc.CallOption) (Stream_FailClient, error) {
	stream, err := c.cc.NewStream(ctx, &Stream_ServiceDesc.Streams[3], "/codepix.transaction.write.Stream/Fail", opts...)
	if err != nil {
		return nil, err
	}
	x := &streamFailClient{stream}
	return x, nil
}

type Stream_FailClient interface {
	Send(*FailRequest) error
	Recv() (*FailReply, error)
	grpc.ClientStream
}

type streamFailClient struct {
	grpc.ClientStream
}

func (x *streamFailClient) Send(m *FailRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *streamFailClient) Recv() (*FailReply, error) {
	m := new(FailReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// StreamServer is the server API for Stream service.
// All implementations must embed UnimplementedStreamServer
// for forward compatibility
type StreamServer interface {
	Start(Stream_StartServer) error
	Confirm(Stream_ConfirmServer) error
	Complete(Stream_CompleteServer) error
	Fail(Stream_FailServer) error
	mustEmbedUnimplementedStreamServer()
}

// UnimplementedStreamServer must be embedded to have forward compatible implementations.
type UnimplementedStreamServer struct {
}

func (UnimplementedStreamServer) Start(Stream_StartServer) error {
	return status.Errorf(codes.Unimplemented, "method Start not implemented")
}
func (UnimplementedStreamServer) Confirm(Stream_ConfirmServer) error {
	return status.Errorf(codes.Unimplemented, "method Confirm not implemented")
}
func (UnimplementedStreamServer) Complete(Stream_CompleteServer) error {
	return status.Errorf(codes.Unimplemented, "method Complete not implemented")
}
func (UnimplementedStreamServer) Fail(Stream_FailServer) error {
	return status.Errorf(codes.Unimplemented, "method Fail not implemented")
}
func (UnimplementedStreamServer) mustEmbedUnimplementedStreamServer() {}

// UnsafeStreamServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StreamServer will
// result in compilation errors.
type UnsafeStreamServer interface {
	mustEmbedUnimplementedStreamServer()
}

func RegisterStreamServer(s grpc.ServiceRegistrar, srv StreamServer) {
	s.RegisterService(&Stream_ServiceDesc, srv)
}

func _Stream_Start_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(StreamServer).Start(&streamStartServer{stream})
}

type Stream_StartServer interface {
	Send(*StartReply) error
	Recv() (*StartRequest, error)
	grpc.ServerStream
}

type streamStartServer struct {
	grpc.ServerStream
}

func (x *streamStartServer) Send(m *StartReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *streamStartServer) Recv() (*StartRequest, error) {
	m := new(StartRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Stream_Confirm_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(StreamServer).Confirm(&streamConfirmServer{stream})
}

type Stream_ConfirmServer interface {
	Send(*ConfirmReply) error
	Recv() (*ConfirmRequest, error)
	grpc.ServerStream
}

type streamConfirmServer struct {
	grpc.ServerStream
}

func (x *streamConfirmServer) Send(m *ConfirmReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *streamConfirmServer) Recv() (*ConfirmRequest, error) {
	m := new(ConfirmRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Stream_Complete_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(StreamServer).Complete(&streamCompleteServer{stream})
}

type Stream_CompleteServer interface {
	Send(*CompleteReply) error
	Recv() (*CompleteRequest, error)
	grpc.ServerStream
}

type streamCompleteServer struct {
	grpc.ServerStream
}

func (x *streamCompleteServer) Send(m *CompleteReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *streamCompleteServer) Recv() (*CompleteRequest, error) {
	m := new(CompleteRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Stream_Fail_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(StreamServer).Fail(&streamFailServer{stream})
}

type Stream_FailServer interface {
	Send(*FailReply) error
	Recv() (*FailRequest, error)
	grpc.ServerStream
}

type streamFailServer struct {
	grpc.ServerStream
}

func (x *streamFailServer) Send(m *FailReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *streamFailServer) Recv() (*FailRequest, error) {
	m := new(FailRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Stream_ServiceDesc is the grpc.ServiceDesc for Stream service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Stream_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "codepix.transaction.write.Stream",
	HandlerType: (*StreamServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Start",
			Handler:       _Stream_Start_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Confirm",
			Handler:       _Stream_Confirm_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Complete",
			Handler:       _Stream_Complete_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Fail",
			Handler:       _Stream_Fail_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/codepix/transaction/write/stream.proto",
}
