// Code generated by protoc-gen-go. DO NOT EDIT.
// sync.proto is a deprecated file.

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

//配置的事件类型
type ConfigEventType int32

const (
	// type
	ConfigEventType__CONFIG_EVENT_TYPE_ ConfigEventType = 0
	// 配置无改变事件,为了兼容以前的事件类型，none设置为负数
	ConfigEventType_CONFIG_EVENT_NONE ConfigEventType = -1
	// 配置同步事件
	ConfigEventType_CONFIG_EVENT_SYNC ConfigEventType = 1
	// 节点配置更改
	ConfigEventType_CONFIG_EVENT_NODE_CONF_CHANGED ConfigEventType = 2
	// 节点注册
	ConfigEventType_CONFIG_EVENT_NODE_REGISTER ConfigEventType = 3
	// 节点状态事件
	ConfigEventType_CONFIG_EVENT_NODE_STATUS ConfigEventType = 4
	// 节点运行资源事件
	ConfigEventType_CONFIG_EVENT_NODE_PROC ConfigEventType = 5
	// 节点退出事件
	ConfigEventType_CONFIG_EVENT_NODE_EXIT ConfigEventType = 6
	// 命令启动事件
	ConfigEventType_CONFIG_EVENT_CMD_START ConfigEventType = 7
	// 命令停止事件
	ConfigEventType_CONFIG_EVENT_CMD_STOP ConfigEventType = 8
	// 命令重启事件
	ConfigEventType_CONFIG_EVENT_CMD_RESTART ConfigEventType = 9
	// 配置操作事件
	ConfigEventType_CONFIG_EVENT_OPERATE ConfigEventType = 10
	// 配置批量操作事件
	ConfigEventType_CONFIG_EVENT_OPERATE_BATCH ConfigEventType = 11
	// 配置回滚事件
	ConfigEventType_CONFIG_OPERATE_ROLLBACK ConfigEventType = 12
)

var ConfigEventType_name = map[int32]string{
	0:  "_CONFIG_EVENT_TYPE_",
	-1: "CONFIG_EVENT_NONE",
	1:  "CONFIG_EVENT_SYNC",
	2:  "CONFIG_EVENT_NODE_CONF_CHANGED",
	3:  "CONFIG_EVENT_NODE_REGISTER",
	4:  "CONFIG_EVENT_NODE_STATUS",
	5:  "CONFIG_EVENT_NODE_PROC",
	6:  "CONFIG_EVENT_NODE_EXIT",
	7:  "CONFIG_EVENT_CMD_START",
	8:  "CONFIG_EVENT_CMD_STOP",
	9:  "CONFIG_EVENT_CMD_RESTART",
	10: "CONFIG_EVENT_OPERATE",
	11: "CONFIG_EVENT_OPERATE_BATCH",
	12: "CONFIG_OPERATE_ROLLBACK",
}

var ConfigEventType_value = map[string]int32{
	"_CONFIG_EVENT_TYPE_":            0,
	"CONFIG_EVENT_NONE":              -1,
	"CONFIG_EVENT_SYNC":              1,
	"CONFIG_EVENT_NODE_CONF_CHANGED": 2,
	"CONFIG_EVENT_NODE_REGISTER":     3,
	"CONFIG_EVENT_NODE_STATUS":       4,
	"CONFIG_EVENT_NODE_PROC":         5,
	"CONFIG_EVENT_NODE_EXIT":         6,
	"CONFIG_EVENT_CMD_START":         7,
	"CONFIG_EVENT_CMD_STOP":          8,
	"CONFIG_EVENT_CMD_RESTART":       9,
	"CONFIG_EVENT_OPERATE":           10,
	"CONFIG_EVENT_OPERATE_BATCH":     11,
	"CONFIG_OPERATE_ROLLBACK":        12,
}

func (x ConfigEventType) String() string {
	return proto.EnumName(ConfigEventType_name, int32(x))
}

func (ConfigEventType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5273b98214de8075, []int{0}
}

//配置变动的接口体
type ConfigEvent struct {
	//请求的事件类型
	EventType ConfigEventType `protobuf:"varint,1,opt,name=event_type,json=eventType,proto3,enum=keeper.ConfigEventType" json:"event_type,omitempty"`
	//数据信息,json格式
	Data                 string   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConfigEvent) Reset()         { *m = ConfigEvent{} }
func (m *ConfigEvent) String() string { return proto.CompactTextString(m) }
func (*ConfigEvent) ProtoMessage()    {}
func (*ConfigEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_5273b98214de8075, []int{0}
}

func (m *ConfigEvent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigEvent.Unmarshal(m, b)
}
func (m *ConfigEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigEvent.Marshal(b, m, deterministic)
}
func (m *ConfigEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigEvent.Merge(m, src)
}
func (m *ConfigEvent) XXX_Size() int {
	return xxx_messageInfo_ConfigEvent.Size(m)
}
func (m *ConfigEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigEvent.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigEvent proto.InternalMessageInfo

func (m *ConfigEvent) GetEventType() ConfigEventType {
	if m != nil {
		return m.EventType
	}
	return ConfigEventType__CONFIG_EVENT_TYPE_
}

func (m *ConfigEvent) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

func init() {
	proto.RegisterEnum("keeper.ConfigEventType", ConfigEventType_name, ConfigEventType_value)
	proto.RegisterType((*ConfigEvent)(nil), "keeper.ConfigEvent")
}

func init() { proto.RegisterFile("sync.proto", fileDescriptor_5273b98214de8075) }

var fileDescriptor_5273b98214de8075 = []byte{
	// 394 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0xd1, 0x6e, 0xd3, 0x30,
	0x14, 0x86, 0xe7, 0xb4, 0x14, 0x7a, 0x86, 0xc0, 0x9c, 0x31, 0x1a, 0xc2, 0x54, 0x4d, 0xbd, 0xaa,
	0xb8, 0xc8, 0x50, 0x91, 0x76, 0x89, 0x94, 0xb8, 0xa6, 0xad, 0x18, 0x49, 0xe4, 0x18, 0x44, 0xb9,
	0x89, 0xb2, 0x60, 0xaa, 0x08, 0x88, 0xa3, 0x36, 0x9a, 0x94, 0xd7, 0xe1, 0x86, 0x57, 0xe0, 0xed,
	0x40, 0x69, 0x53, 0x69, 0x6d, 0xe2, 0x1b, 0x1f, 0x9d, 0xef, 0xff, 0x8f, 0x6d, 0xf9, 0x07, 0xd8,
	0x94, 0x59, 0x62, 0xe7, 0x6b, 0x5d, 0x68, 0xec, 0xfd, 0x50, 0x2a, 0x57, 0x6b, 0xeb, 0x62, 0xa5,
	0xf5, 0xea, 0xa7, 0xba, 0x8a, 0xf3, 0xf4, 0x2a, 0xce, 0x32, 0x5d, 0xc4, 0x45, 0xaa, 0xb3, 0xcd,
	0x4e, 0x35, 0x5a, 0xc2, 0x29, 0xd3, 0xd9, 0xf7, 0x74, 0xc5, 0xef, 0x54, 0x56, 0xe0, 0x35, 0x80,
	0xaa, 0x8a, 0xa8, 0x28, 0x73, 0x65, 0x92, 0x4b, 0x32, 0x7e, 0x32, 0x19, 0xd8, 0xbb, 0x49, 0xf6,
	0x3d, 0xa1, 0x2c, 0x73, 0x25, 0xfa, 0x6a, 0x5f, 0x22, 0x42, 0xf7, 0x5b, 0x5c, 0xc4, 0xa6, 0x71,
	0x49, 0xc6, 0x7d, 0xb1, 0xad, 0x5f, 0xff, 0xe9, 0xc0, 0xd3, 0x23, 0x0b, 0x0e, 0xe0, 0x2c, 0x62,
	0xbe, 0xf7, 0x7e, 0x31, 0x8b, 0xf8, 0x67, 0xee, 0xc9, 0x48, 0x2e, 0x03, 0x1e, 0xd1, 0x13, 0x1c,
	0xc2, 0xb3, 0x83, 0xbe, 0xe7, 0x7b, 0x9c, 0xfe, 0xdb, 0x2f, 0x82, 0xe7, 0x47, 0x3c, 0x5c, 0x7a,
	0x8c, 0x12, 0x1c, 0xc1, 0xf0, 0xc8, 0x36, 0xe5, 0xdb, 0x03, 0x22, 0x36, 0x77, 0xbc, 0x19, 0x9f,
	0x52, 0x03, 0x87, 0x60, 0x35, 0x35, 0x82, 0xcf, 0x16, 0xa1, 0xe4, 0x82, 0x76, 0xf0, 0x02, 0xcc,
	0x26, 0x0f, 0xa5, 0x23, 0x3f, 0x85, 0xb4, 0x8b, 0x16, 0xbc, 0x68, 0xd2, 0x40, 0xf8, 0x8c, 0x3e,
	0x68, 0x67, 0xfc, 0xcb, 0x42, 0xd2, 0x5e, 0x83, 0xb1, 0x8f, 0xd3, 0x6a, 0xa8, 0x90, 0xf4, 0x21,
	0xbe, 0x84, 0xf3, 0x16, 0xe6, 0x07, 0xf4, 0x51, 0xe3, 0x32, 0x15, 0x12, 0x7c, 0x67, 0xec, 0xa3,
	0x09, 0xcf, 0x0f, 0xa8, 0x1f, 0x70, 0xe1, 0x48, 0x4e, 0xa1, 0xf1, 0xc8, 0x9a, 0x44, 0xae, 0x23,
	0xd9, 0x9c, 0x9e, 0xe2, 0x2b, 0x18, 0xd4, 0x7c, 0x4f, 0x84, 0x7f, 0x73, 0xe3, 0x3a, 0xec, 0x03,
	0x7d, 0x3c, 0x79, 0x07, 0xdd, 0xb0, 0xcc, 0x12, 0xbc, 0xae, 0xf7, 0xb3, 0x96, 0x1f, 0xb7, 0xda,
	0x9a, 0xa3, 0x93, 0x31, 0x79, 0x43, 0xdc, 0x09, 0x60, 0xa2, 0x7f, 0xd9, 0x1b, 0xb5, 0xbe, 0x4b,
	0x13, 0x55, 0xeb, 0xdc, 0x3a, 0x58, 0x41, 0x95, 0xb3, 0x39, 0x09, 0xc8, 0x57, 0x23, 0xbf, 0xfd,
	0x4b, 0xc8, 0x6f, 0xa3, 0xc3, 0x02, 0xf7, 0xb6, 0xb7, 0xcd, 0xdf, 0xdb, 0xff, 0x01, 0x00, 0x00,
	0xff, 0xff, 0x41, 0xf0, 0x26, 0xac, 0xb3, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SyncClient is the client API for Sync service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SyncClient interface {
	//获取配置接口,如果请求的事件为CONFIG_EVENT_NONE且keeper中的配置没有发生改变，会阻塞n秒
	Sync(ctx context.Context, opts ...grpc.CallOption) (Sync_SyncClient, error)
}

type syncClient struct {
	cc *grpc.ClientConn
}

func NewSyncClient(cc *grpc.ClientConn) SyncClient {
	return &syncClient{cc}
}

func (c *syncClient) Sync(ctx context.Context, opts ...grpc.CallOption) (Sync_SyncClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Sync_serviceDesc.Streams[0], "/keeper.Sync/Sync", opts...)
	if err != nil {
		return nil, err
	}
	x := &syncSyncClient{stream}
	return x, nil
}

type Sync_SyncClient interface {
	Send(*ConfigEvent) error
	Recv() (*ConfigEvent, error)
	grpc.ClientStream
}

type syncSyncClient struct {
	grpc.ClientStream
}

func (x *syncSyncClient) Send(m *ConfigEvent) error {
	return x.ClientStream.SendMsg(m)
}

func (x *syncSyncClient) Recv() (*ConfigEvent, error) {
	m := new(ConfigEvent)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SyncServer is the server API for Sync service.
type SyncServer interface {
	//获取配置接口,如果请求的事件为CONFIG_EVENT_NONE且keeper中的配置没有发生改变，会阻塞n秒
	Sync(Sync_SyncServer) error
}

// UnimplementedSyncServer can be embedded to have forward compatible implementations.
type UnimplementedSyncServer struct {
}

func (*UnimplementedSyncServer) Sync(srv Sync_SyncServer) error {
	return status.Errorf(codes.Unimplemented, "method Sync not implemented")
}

func RegisterSyncServer(s *grpc.Server, srv SyncServer) {
	s.RegisterService(&_Sync_serviceDesc, srv)
}

func _Sync_Sync_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SyncServer).Sync(&syncSyncServer{stream})
}

type Sync_SyncServer interface {
	Send(*ConfigEvent) error
	Recv() (*ConfigEvent, error)
	grpc.ServerStream
}

type syncSyncServer struct {
	grpc.ServerStream
}

func (x *syncSyncServer) Send(m *ConfigEvent) error {
	return x.ServerStream.SendMsg(m)
}

func (x *syncSyncServer) Recv() (*ConfigEvent, error) {
	m := new(ConfigEvent)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Sync_serviceDesc = grpc.ServiceDesc{
	ServiceName: "keeper.Sync",
	HandlerType: (*SyncServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Sync",
			Handler:       _Sync_Sync_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "sync.proto",
}
