package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/golang/protobuf/ptypes/empty"
import _ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
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
	Name                 string   `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HelloRequest) Reset()         { *m = HelloRequest{} }
func (m *HelloRequest) String() string { return proto.CompactTextString(m) }
func (*HelloRequest) ProtoMessage()    {}
func (*HelloRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_myapp_8ad0b6653764be02, []int{0}
}
func (m *HelloRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HelloRequest.Unmarshal(m, b)
}
func (m *HelloRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HelloRequest.Marshal(b, m, deterministic)
}
func (dst *HelloRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HelloRequest.Merge(dst, src)
}
func (m *HelloRequest) XXX_Size() int {
	return xxx_messageInfo_HelloRequest.Size(m)
}
func (m *HelloRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_HelloRequest.DiscardUnknown(m)
}

var xxx_messageInfo_HelloRequest proto.InternalMessageInfo

func (m *HelloRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type HelloResponse struct {
	Message              string   `protobuf:"bytes,1,opt,name=Message" json:"Message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HelloResponse) Reset()         { *m = HelloResponse{} }
func (m *HelloResponse) String() string { return proto.CompactTextString(m) }
func (*HelloResponse) ProtoMessage()    {}
func (*HelloResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_myapp_8ad0b6653764be02, []int{1}
}
func (m *HelloResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HelloResponse.Unmarshal(m, b)
}
func (m *HelloResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HelloResponse.Marshal(b, m, deterministic)
}
func (dst *HelloResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HelloResponse.Merge(dst, src)
}
func (m *HelloResponse) XXX_Size() int {
	return xxx_messageInfo_HelloResponse.Size(m)
}
func (m *HelloResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_HelloResponse.DiscardUnknown(m)
}

var xxx_messageInfo_HelloResponse proto.InternalMessageInfo

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

func init() { proto.RegisterFile("myapp.proto", fileDescriptor_myapp_8ad0b6653764be02) }

var fileDescriptor_myapp_8ad0b6653764be02 = []byte{
	// 263 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x8f, 0xc1, 0x4a, 0x03, 0x31,
	0x18, 0x84, 0xd9, 0x62, 0x5b, 0x4d, 0xf5, 0x12, 0x54, 0xca, 0xea, 0x41, 0x72, 0x10, 0x15, 0x9b,
	0x80, 0xbd, 0x79, 0xd3, 0x93, 0x07, 0xeb, 0xa1, 0xe0, 0xc5, 0x5b, 0x22, 0xbf, 0xe9, 0xe2, 0x6e,
	0xfe, 0xdf, 0xfd, 0x53, 0x64, 0xaf, 0x3e, 0x81, 0xe8, 0xa3, 0xf9, 0x0a, 0x3e, 0x88, 0x34, 0xbb,
	0x05, 0xd1, 0x53, 0x32, 0x33, 0xdf, 0x24, 0x8c, 0x18, 0x55, 0x8d, 0x25, 0xd2, 0x54, 0x63, 0x44,
	0x39, 0x4c, 0x82, 0x5c, 0x7e, 0xe0, 0x11, 0x7d, 0x09, 0x26, 0xd9, 0x6e, 0xf9, 0x64, 0xa0, 0xa2,
	0xd8, 0xb4, 0x54, 0x7e, 0xd8, 0x85, 0x96, 0x0a, 0x63, 0x43, 0xc0, 0x68, 0x63, 0x81, 0x81, 0xbb,
	0xf4, 0x3c, 0x1d, 0x8f, 0x13, 0x0f, 0x61, 0xc2, 0xaf, 0xd6, 0x7b, 0xa8, 0x0d, 0x52, 0x22, 0xfe,
	0xd3, 0x4a, 0x89, 0xed, 0x1b, 0x28, 0x4b, 0x9c, 0xc3, 0xcb, 0x12, 0x38, 0x4a, 0x29, 0x36, 0xee,
	0x6c, 0x05, 0xe3, 0xec, 0x28, 0x3b, 0xd9, 0x9a, 0xa7, 0xbb, 0x3a, 0x15, 0x3b, 0x1d, 0xc3, 0x84,
	0x81, 0x41, 0x8e, 0xc5, 0x70, 0x06, 0xcc, 0xd6, 0xaf, 0xb9, 0xb5, 0xbc, 0xb8, 0x17, 0xfd, 0xd9,
	0x6a, 0x82, 0xbc, 0x15, 0xfd, 0xd4, 0x91, 0x7b, 0xba, 0xdb, 0xa4, 0x7f, 0xff, 0x93, 0xef, 0xff,
	0xb5, 0xdb, 0xa7, 0xd5, 0xee, 0xdb, 0xd7, 0xf7, 0x67, 0x6f, 0xa4, 0x06, 0x66, 0xb1, 0xf2, 0x2f,
	0xb3, 0xb3, 0xf7, 0x5e, 0x76, 0x3d, 0xfd, 0xb8, 0xda, 0x54, 0x03, 0x93, 0x2a, 0x0f, 0xc7, 0xbe,
	0x88, 0xa5, 0x75, 0xba, 0x6a, 0x22, 0x3c, 0x73, 0xa1, 0x03, 0x44, 0xe3, 0x91, 0x16, 0x50, 0xb3,
	0xf1, 0xd8, 0x42, 0x86, 0x9c, 0x1b, 0xa4, 0x85, 0xd3, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x00,
	0x48, 0x7a, 0x59, 0x62, 0x01, 0x00, 0x00,
}
