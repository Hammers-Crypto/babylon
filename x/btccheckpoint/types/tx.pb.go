// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: babylon/btccheckpoint/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Consider we have a Merkle tree with following structure:
//            ROOT
//           /    \
//      H1234      H5555
//     /     \       \
//   H12     H34      H55
//  /  \    /  \     /
// H1  H2  H3  H4  H5
// L1  L2  L3  L4  L5
// To prove L3 was part of ROOT we need:
// - btc_transaction_index = 2 which in binary is 010
// (where 0 means going left, 1 means going right in the tree)
// - merkle_nodes we'd have H4 || H12 || H5555
// By looking at 010 we would know that H4 is a right sibling,
// H12 is left, H5555 is right again.
type BtcSpvProof struct {
	// Valid bitcoin transaction containing OP_RETURN opcode.
	BtcTransaction []byte `protobuf:"bytes,1,opt,name=btc_transaction,json=btcTransaction,proto3" json:"btc_transaction,omitempty"`
	// Index of transaction within the block. Index is needed to determine if
	// currently hashed node is left or right.
	BtcTransactionIndex uint32 `protobuf:"varint,2,opt,name=btc_transaction_index,json=btcTransactionIndex,proto3" json:"btc_transaction_index,omitempty"`
	// List of concatenated intermediate merkele tree nodes, without root node and leaf node
	// against which we calculate the proof.
	// Each node has 32 byte length.
	// Example proof can look like: 32_bytes_of_node1 || 32_bytes_of_node2 ||  32_bytes_of_node3
	// so the length of the proof will always be divisible by 32.
	MerkleNodes []byte `protobuf:"bytes,3,opt,name=merkle_nodes,json=merkleNodes,proto3" json:"merkle_nodes,omitempty"`
	// Valid btc header which confirms btc_transaction.
	// Should have exactly 80 bytes
	ConfirmingBtcHeader []byte `protobuf:"bytes,4,opt,name=confirming_btc_header,json=confirmingBtcHeader,proto3" json:"confirming_btc_header,omitempty"`
}

func (m *BtcSpvProof) Reset()         { *m = BtcSpvProof{} }
func (m *BtcSpvProof) String() string { return proto.CompactTextString(m) }
func (*BtcSpvProof) ProtoMessage()    {}
func (*BtcSpvProof) Descriptor() ([]byte, []int) {
	return fileDescriptor_aeec89810b39ea83, []int{0}
}
func (m *BtcSpvProof) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BtcSpvProof) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BtcSpvProof.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BtcSpvProof) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BtcSpvProof.Merge(m, src)
}
func (m *BtcSpvProof) XXX_Size() int {
	return m.Size()
}
func (m *BtcSpvProof) XXX_DiscardUnknown() {
	xxx_messageInfo_BtcSpvProof.DiscardUnknown(m)
}

var xxx_messageInfo_BtcSpvProof proto.InternalMessageInfo

func (m *BtcSpvProof) GetBtcTransaction() []byte {
	if m != nil {
		return m.BtcTransaction
	}
	return nil
}

func (m *BtcSpvProof) GetBtcTransactionIndex() uint32 {
	if m != nil {
		return m.BtcTransactionIndex
	}
	return 0
}

func (m *BtcSpvProof) GetMerkleNodes() []byte {
	if m != nil {
		return m.MerkleNodes
	}
	return nil
}

func (m *BtcSpvProof) GetConfirmingBtcHeader() []byte {
	if m != nil {
		return m.ConfirmingBtcHeader
	}
	return nil
}

type InsertBtcSpvProofRequest struct {
	Proofs []*BtcSpvProof `protobuf:"bytes,1,rep,name=proofs,proto3" json:"proofs,omitempty"`
}

func (m *InsertBtcSpvProofRequest) Reset()         { *m = InsertBtcSpvProofRequest{} }
func (m *InsertBtcSpvProofRequest) String() string { return proto.CompactTextString(m) }
func (*InsertBtcSpvProofRequest) ProtoMessage()    {}
func (*InsertBtcSpvProofRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_aeec89810b39ea83, []int{1}
}
func (m *InsertBtcSpvProofRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InsertBtcSpvProofRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InsertBtcSpvProofRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InsertBtcSpvProofRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InsertBtcSpvProofRequest.Merge(m, src)
}
func (m *InsertBtcSpvProofRequest) XXX_Size() int {
	return m.Size()
}
func (m *InsertBtcSpvProofRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_InsertBtcSpvProofRequest.DiscardUnknown(m)
}

var xxx_messageInfo_InsertBtcSpvProofRequest proto.InternalMessageInfo

func (m *InsertBtcSpvProofRequest) GetProofs() []*BtcSpvProof {
	if m != nil {
		return m.Proofs
	}
	return nil
}

type InsertBtcSpvProofResponse struct {
}

func (m *InsertBtcSpvProofResponse) Reset()         { *m = InsertBtcSpvProofResponse{} }
func (m *InsertBtcSpvProofResponse) String() string { return proto.CompactTextString(m) }
func (*InsertBtcSpvProofResponse) ProtoMessage()    {}
func (*InsertBtcSpvProofResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_aeec89810b39ea83, []int{2}
}
func (m *InsertBtcSpvProofResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InsertBtcSpvProofResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InsertBtcSpvProofResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InsertBtcSpvProofResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InsertBtcSpvProofResponse.Merge(m, src)
}
func (m *InsertBtcSpvProofResponse) XXX_Size() int {
	return m.Size()
}
func (m *InsertBtcSpvProofResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_InsertBtcSpvProofResponse.DiscardUnknown(m)
}

var xxx_messageInfo_InsertBtcSpvProofResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*BtcSpvProof)(nil), "babylonchain.babylon.btccheckpoint.BtcSpvProof")
	proto.RegisterType((*InsertBtcSpvProofRequest)(nil), "babylonchain.babylon.btccheckpoint.InsertBtcSpvProofRequest")
	proto.RegisterType((*InsertBtcSpvProofResponse)(nil), "babylonchain.babylon.btccheckpoint.InsertBtcSpvProofResponse")
}

func init() { proto.RegisterFile("babylon/btccheckpoint/tx.proto", fileDescriptor_aeec89810b39ea83) }

var fileDescriptor_aeec89810b39ea83 = []byte{
	// 340 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x52, 0x3d, 0x4f, 0x02, 0x31,
	0x18, 0xa6, 0x62, 0x18, 0x0a, 0x6a, 0x2c, 0x31, 0x39, 0x35, 0x69, 0xf0, 0x16, 0x99, 0xee, 0x12,
	0x8c, 0x9b, 0x2e, 0x2c, 0xca, 0xe0, 0x47, 0xd0, 0xc9, 0xe5, 0x72, 0x2d, 0x85, 0x6b, 0x80, 0xf6,
	0x6c, 0x5f, 0x0c, 0xfc, 0x0b, 0x46, 0xff, 0x8e, 0x9b, 0x23, 0xa3, 0xa3, 0x81, 0x3f, 0x62, 0xee,
	0xc0, 0x70, 0xa0, 0x46, 0xe3, 0xf8, 0x7c, 0xbc, 0xcf, 0xd3, 0xb7, 0x79, 0x31, 0x65, 0x21, 0x1b,
	0xf5, 0xb4, 0xf2, 0x19, 0x70, 0x1e, 0x09, 0xde, 0x8d, 0xb5, 0x54, 0xe0, 0xc3, 0xd0, 0x8b, 0x8d,
	0x06, 0x4d, 0xdc, 0x85, 0xce, 0xa3, 0x50, 0x2a, 0x6f, 0x01, 0xbc, 0x15, 0xb3, 0xfb, 0x82, 0x70,
	0xb1, 0x0e, 0xfc, 0x2e, 0x7e, 0xba, 0x35, 0x5a, 0xb7, 0xc9, 0x31, 0xde, 0x61, 0xc0, 0x03, 0x30,
	0xa1, 0xb2, 0x21, 0x07, 0xa9, 0x95, 0x83, 0x2a, 0xa8, 0x5a, 0x6a, 0x6e, 0x33, 0xe0, 0xf7, 0x4b,
	0x96, 0xd4, 0xf0, 0xde, 0x9a, 0x31, 0x90, 0xaa, 0x25, 0x86, 0xce, 0x46, 0x05, 0x55, 0xb7, 0x9a,
	0xe5, 0x55, 0x7b, 0x23, 0x91, 0xc8, 0x11, 0x2e, 0xf5, 0x85, 0xe9, 0xf6, 0x44, 0xa0, 0x74, 0x4b,
	0x58, 0x27, 0x9f, 0x26, 0x17, 0xe7, 0xdc, 0x75, 0x42, 0x25, 0xb1, 0x5c, 0xab, 0xb6, 0x34, 0x7d,
	0xa9, 0x3a, 0x41, 0xd2, 0x10, 0x89, 0xb0, 0x25, 0x8c, 0xb3, 0x99, 0x7a, 0xcb, 0x4b, 0xb1, 0x0e,
	0xfc, 0x32, 0x95, 0x5c, 0x8e, 0x9d, 0x86, 0xb2, 0xc2, 0x40, 0x66, 0x91, 0xa6, 0x78, 0x1c, 0x08,
	0x0b, 0xe4, 0x02, 0x17, 0xe2, 0x04, 0x5b, 0x07, 0x55, 0xf2, 0xd5, 0x62, 0xcd, 0xf7, 0x7e, 0xff,
	0x14, 0x2f, 0x9b, 0xb3, 0x18, 0x77, 0x0f, 0xf1, 0xfe, 0x37, 0x25, 0x36, 0xd6, 0xca, 0x8a, 0xda,
	0x33, 0xc2, 0xf9, 0x2b, 0xdb, 0x21, 0x63, 0x84, 0x77, 0xbf, 0xb8, 0xc8, 0xd9, 0x5f, 0x3a, 0x7f,
	0xda, 0xe0, 0xe0, 0xfc, 0x9f, 0xd3, 0xf3, 0xa7, 0xd5, 0x6f, 0x5e, 0xa7, 0x14, 0x4d, 0xa6, 0x14,
	0xbd, 0x4f, 0x29, 0x1a, 0xcf, 0x68, 0x6e, 0x32, 0xa3, 0xb9, 0xb7, 0x19, 0xcd, 0x3d, 0x9c, 0x76,
	0x24, 0x44, 0x03, 0xe6, 0x71, 0xdd, 0xf7, 0xb3, 0x15, 0x9f, 0xc0, 0x1f, 0xae, 0x1f, 0xd6, 0x28,
	0x16, 0x96, 0x15, 0xd2, 0xe3, 0x3a, 0xf9, 0x08, 0x00, 0x00, 0xff, 0xff, 0x9f, 0x0a, 0x2a, 0xff,
	0x7e, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	InsertBtcSpvProof(ctx context.Context, in *InsertBtcSpvProofRequest, opts ...grpc.CallOption) (*InsertBtcSpvProofResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) InsertBtcSpvProof(ctx context.Context, in *InsertBtcSpvProofRequest, opts ...grpc.CallOption) (*InsertBtcSpvProofResponse, error) {
	out := new(InsertBtcSpvProofResponse)
	err := c.cc.Invoke(ctx, "/babylonchain.babylon.btccheckpoint.Msg/InsertBtcSpvProof", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	InsertBtcSpvProof(context.Context, *InsertBtcSpvProofRequest) (*InsertBtcSpvProofResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) InsertBtcSpvProof(ctx context.Context, req *InsertBtcSpvProofRequest) (*InsertBtcSpvProofResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InsertBtcSpvProof not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_InsertBtcSpvProof_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertBtcSpvProofRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).InsertBtcSpvProof(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/babylonchain.babylon.btccheckpoint.Msg/InsertBtcSpvProof",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).InsertBtcSpvProof(ctx, req.(*InsertBtcSpvProofRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "babylonchain.babylon.btccheckpoint.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InsertBtcSpvProof",
			Handler:    _Msg_InsertBtcSpvProof_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "babylon/btccheckpoint/tx.proto",
}

func (m *BtcSpvProof) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BtcSpvProof) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BtcSpvProof) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ConfirmingBtcHeader) > 0 {
		i -= len(m.ConfirmingBtcHeader)
		copy(dAtA[i:], m.ConfirmingBtcHeader)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ConfirmingBtcHeader)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.MerkleNodes) > 0 {
		i -= len(m.MerkleNodes)
		copy(dAtA[i:], m.MerkleNodes)
		i = encodeVarintTx(dAtA, i, uint64(len(m.MerkleNodes)))
		i--
		dAtA[i] = 0x1a
	}
	if m.BtcTransactionIndex != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.BtcTransactionIndex))
		i--
		dAtA[i] = 0x10
	}
	if len(m.BtcTransaction) > 0 {
		i -= len(m.BtcTransaction)
		copy(dAtA[i:], m.BtcTransaction)
		i = encodeVarintTx(dAtA, i, uint64(len(m.BtcTransaction)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *InsertBtcSpvProofRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InsertBtcSpvProofRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InsertBtcSpvProofRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Proofs) > 0 {
		for iNdEx := len(m.Proofs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Proofs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *InsertBtcSpvProofResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InsertBtcSpvProofResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InsertBtcSpvProofResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *BtcSpvProof) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.BtcTransaction)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.BtcTransactionIndex != 0 {
		n += 1 + sovTx(uint64(m.BtcTransactionIndex))
	}
	l = len(m.MerkleNodes)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ConfirmingBtcHeader)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *InsertBtcSpvProofRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Proofs) > 0 {
		for _, e := range m.Proofs {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	return n
}

func (m *InsertBtcSpvProofResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *BtcSpvProof) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BtcSpvProof: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BtcSpvProof: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BtcTransaction", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BtcTransaction = append(m.BtcTransaction[:0], dAtA[iNdEx:postIndex]...)
			if m.BtcTransaction == nil {
				m.BtcTransaction = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BtcTransactionIndex", wireType)
			}
			m.BtcTransactionIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BtcTransactionIndex |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MerkleNodes", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MerkleNodes = append(m.MerkleNodes[:0], dAtA[iNdEx:postIndex]...)
			if m.MerkleNodes == nil {
				m.MerkleNodes = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConfirmingBtcHeader", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ConfirmingBtcHeader = append(m.ConfirmingBtcHeader[:0], dAtA[iNdEx:postIndex]...)
			if m.ConfirmingBtcHeader == nil {
				m.ConfirmingBtcHeader = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *InsertBtcSpvProofRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: InsertBtcSpvProofRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InsertBtcSpvProofRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proofs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Proofs = append(m.Proofs, &BtcSpvProof{})
			if err := m.Proofs[len(m.Proofs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *InsertBtcSpvProofResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: InsertBtcSpvProofResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InsertBtcSpvProofResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
