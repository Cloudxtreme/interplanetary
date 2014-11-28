// Code generated by protoc-gen-gogo.
// source: message.proto
// DO NOT EDIT!

/*
Package bitswap_message_pb is a generated protocol buffer package.

It is generated from these files:
	message.proto

It has these top-level messages:
	Message
*/
package bitswap_message_pb

import proto "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/gogoprotobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type Message struct {
	Wantlist         []string `protobuf:"bytes,1,rep,name=wantlist" json:"wantlist,omitempty"`
	Blocks           [][]byte `protobuf:"bytes,2,rep,name=blocks" json:"blocks,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}

func (m *Message) GetWantlist() []string {
	if m != nil {
		return m.Wantlist
	}
	return nil
}

func (m *Message) GetBlocks() [][]byte {
	if m != nil {
		return m.Blocks
	}
	return nil
}

func init() {
}