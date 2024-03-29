// Code generated by protoc-gen-gogo.
// source: diagnostics.proto
// DO NOT EDIT!

/*
Package diagnostics_pb is a generated protocol buffer package.

It is generated from these files:
	diagnostics.proto

It has these top-level messages:
	Message
*/
package diagnostics_pb

import proto "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/gogoprotobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type Message struct {
	DiagID           *string `protobuf:"bytes,1,req" json:"DiagID,omitempty"`
	Data             []byte  `protobuf:"bytes,2,opt" json:"Data,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}

func (m *Message) GetDiagID() string {
	if m != nil && m.DiagID != nil {
		return *m.DiagID
	}
	return ""
}

func (m *Message) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
}
