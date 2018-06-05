package main

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/golang/protobuf/ptypes/empty"
import _ "google.golang.org/genproto/googleapis/api/annotations"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type HelloRequest struct {
	Name string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
}

func (m *HelloRequest) Reset()                    { *m = HelloRequest{} }
func (m *HelloRequest) String() string            { return proto.CompactTextString(m) }
func (*HelloRequest) ProtoMessage()               {}
func (*HelloRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *HelloRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type HelloResponse struct {
	Message string `protobuf:"bytes,1,opt,name=Message" json:"Message,omitempty"`
}

func (m *HelloResponse) Reset()                    { *m = HelloResponse{} }
func (m *HelloResponse) String() string            { return proto.CompactTextString(m) }
func (*HelloResponse) ProtoMessage()               {}
func (*HelloResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *HelloResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*HelloRequest)(nil), "myapppb.HelloRequest")
	proto.RegisterType((*HelloResponse)(nil), "myapppb.HelloResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Myapp service

type MyappClient interface {
	Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
}

type myappClient struct {
	cc *grpc.ClientConn
}

func NewMyappClient(cc *grpc.ClientConn) MyappClient {
	return &myappClient{cc}
}

func (c *myappClient) Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error) {
	out := new(HelloResponse)
	err := grpc.Invoke(ctx, "/myapppb.Myapp/Hello", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Myapp service

type MyappServer interface {
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
}

func RegisterMyappServer(s *grpc.Server, srv MyappServer) {
	s.RegisterService(&_Myapp_serviceDesc, srv)
}

func _Myapp_Hello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MyappServer).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/myapppb.Myapp/Hello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MyappServer).Hello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Myapp_serviceDesc = grpc.ServiceDesc{
	ServiceName: "myapppb.Myapp",
	HandlerType: (*MyappServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hello",
			Handler:    _Myapp_Hello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "myapp.proto",
}

func init() { proto.RegisterFile("myapp.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 232 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x4f, 0xbd, 0x4e, 0x03, 0x31,
	0x0c, 0xd6, 0x21, 0xda, 0x8a, 0x14, 0x06, 0x22, 0x81, 0xaa, 0x83, 0x01, 0x65, 0x40, 0x85, 0x21,
	0x91, 0x60, 0x63, 0x64, 0x82, 0xa1, 0x0c, 0x37, 0xb2, 0x25, 0x92, 0x49, 0x4f, 0x5c, 0x62, 0x53,
	0xbb, 0xc3, 0xad, 0xbc, 0x02, 0x8f, 0xc6, 0x2b, 0xf0, 0x20, 0xa8, 0xb9, 0xab, 0x84, 0xd8, 0xfc,
	0xfd, 0xf8, 0xf3, 0x67, 0x35, 0x4f, 0xbd, 0x27, 0xb2, 0xb4, 0x41, 0x41, 0x3d, 0x2b, 0x80, 0x42,
	0x7d, 0x11, 0x11, 0x63, 0x07, 0xae, 0xd0, 0x61, 0xfb, 0xe6, 0x20, 0x91, 0xf4, 0x83, 0xab, 0xbe,
	0x1c, 0x45, 0x4f, 0xad, 0xf3, 0x39, 0xa3, 0x78, 0x69, 0x31, 0xf3, 0xa0, 0x1a, 0xa3, 0x8e, 0x9f,
	0xa0, 0xeb, 0xb0, 0x81, 0x8f, 0x2d, 0xb0, 0x68, 0xad, 0x0e, 0x5f, 0x7c, 0x82, 0x45, 0x75, 0x55,
	0x2d, 0x8f, 0x9a, 0x32, 0x9b, 0x1b, 0x75, 0x32, 0x7a, 0x98, 0x30, 0x33, 0xe8, 0x85, 0x9a, 0xad,
	0x80, 0xd9, 0xc7, 0xbd, 0x6f, 0x0f, 0xef, 0x1a, 0x35, 0x59, 0xed, 0x4a, 0xe9, 0x67, 0x35, 0x29,
	0x3b, 0xfa, 0xcc, 0x8e, 0x2d, 0xed, 0xdf, 0x3b, 0xf5, 0xf9, 0x7f, 0x7a, 0x88, 0x36, 0xa7, 0x9f,
	0xdf, 0x3f, 0x5f, 0x07, 0x73, 0x33, 0x75, 0xeb, 0x1d, 0xff, 0x50, 0xdd, 0x3e, 0x2e, 0x5f, 0xaf,
	0x63, 0x2b, 0x9d, 0x0f, 0x36, 0xf5, 0x02, 0xef, 0xdc, 0xda, 0x0c, 0xe2, 0x22, 0xd2, 0x1a, 0x36,
	0xec, 0x22, 0xba, 0x92, 0xe4, 0x28, 0x84, 0x69, 0xf9, 0xe9, 0xfe, 0x37, 0x00, 0x00, 0xff, 0xff,
	0xd1, 0x1c, 0xd5, 0x1e, 0x26, 0x01, 0x00, 0x00,
}
