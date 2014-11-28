// Code generated by protoc-gen-gogo.
// source: merkledag.proto
// DO NOT EDIT!

/*
	Package merkledag_pb is a generated protocol buffer package.

	It is generated from these files:
		merkledag.proto

	It has these top-level messages:
		PBLink
		PBNode
*/
package merkledag_pb

import proto "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/gogoprotobuf/proto"
import math "math"

// discarding unused import gogoproto "code.google.com/p/gogoprotobuf/gogoproto/gogo.pb"

import io "io"
import fmt "fmt"
import code_google_com_p_gogoprotobuf_proto "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/gogoprotobuf/proto"

import fmt1 "fmt"
import strings "strings"
import reflect "reflect"

import fmt2 "fmt"
import strings1 "strings"
import code_google_com_p_gogoprotobuf_proto1 "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/gogoprotobuf/proto"
import sort "sort"
import strconv "strconv"
import reflect1 "reflect"

import fmt3 "fmt"
import bytes "bytes"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

// An IPFS MerkleDAG Link
type PBLink struct {
	// multihash of the target object
	Hash []byte `protobuf:"bytes,1,opt" json:"Hash,omitempty"`
	// utf string name. should be unique per object
	Name *string `protobuf:"bytes,2,opt" json:"Name,omitempty"`
	// cumulative size of target object
	Tsize            *uint64 `protobuf:"varint,3,opt" json:"Tsize,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *PBLink) Reset()      { *m = PBLink{} }
func (*PBLink) ProtoMessage() {}

func (m *PBLink) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *PBLink) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *PBLink) GetTsize() uint64 {
	if m != nil && m.Tsize != nil {
		return *m.Tsize
	}
	return 0
}

// An IPFS MerkleDAG Node
type PBNode struct {
	// refs to other objects
	Links []*PBLink `protobuf:"bytes,2,rep" json:"Links,omitempty"`
	// opaque user data
	Data             []byte `protobuf:"bytes,1,opt" json:"Data,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *PBNode) Reset()      { *m = PBNode{} }
func (*PBNode) ProtoMessage() {}

func (m *PBNode) GetLinks() []*PBLink {
	if m != nil {
		return m.Links
	}
	return nil
}

func (m *PBNode) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
}
func (m *PBLink) Unmarshal(data []byte) error {
	l := len(data)
	index := 0
	for index < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if index >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[index]
			index++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Hash = append(m.Hash, data[index:postIndex]...)
			index = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + int(stringLen)
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(data[index:postIndex])
			m.Name = &s
			index = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tsize", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				v |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Tsize = &v
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			index -= sizeOfWire
			skippy, err := code_google_com_p_gogoprotobuf_proto.Skip(data[index:])
			if err != nil {
				return err
			}
			if (index + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)
			index += skippy
		}
	}
	return nil
}
func (m *PBNode) Unmarshal(data []byte) error {
	l := len(data)
	index := 0
	for index < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if index >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[index]
			index++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Links", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Links = append(m.Links, &PBLink{})
			m.Links[len(m.Links)-1].Unmarshal(data[index:postIndex])
			index = postIndex
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data, data[index:postIndex]...)
			index = postIndex
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			index -= sizeOfWire
			skippy, err := code_google_com_p_gogoprotobuf_proto.Skip(data[index:])
			if err != nil {
				return err
			}
			if (index + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)
			index += skippy
		}
	}
	return nil
}
func (this *PBLink) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&PBLink{`,
		`Hash:` + valueToStringMerkledag(this.Hash) + `,`,
		`Name:` + valueToStringMerkledag(this.Name) + `,`,
		`Tsize:` + valueToStringMerkledag(this.Tsize) + `,`,
		`XXX_unrecognized:` + fmt1.Sprintf("%v", this.XXX_unrecognized) + `,`,
		`}`,
	}, "")
	return s
}
func (this *PBNode) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&PBNode{`,
		`Links:` + strings.Replace(fmt1.Sprintf("%v", this.Links), "PBLink", "PBLink", 1) + `,`,
		`Data:` + valueToStringMerkledag(this.Data) + `,`,
		`XXX_unrecognized:` + fmt1.Sprintf("%v", this.XXX_unrecognized) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringMerkledag(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt1.Sprintf("*%v", pv)
}
func (m *PBLink) Size() (n int) {
	var l int
	_ = l
	if m.Hash != nil {
		l = len(m.Hash)
		n += 1 + l + sovMerkledag(uint64(l))
	}
	if m.Name != nil {
		l = len(*m.Name)
		n += 1 + l + sovMerkledag(uint64(l))
	}
	if m.Tsize != nil {
		n += 1 + sovMerkledag(uint64(*m.Tsize))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}
func (m *PBNode) Size() (n int) {
	var l int
	_ = l
	if len(m.Links) > 0 {
		for _, e := range m.Links {
			l = e.Size()
			n += 1 + l + sovMerkledag(uint64(l))
		}
	}
	if m.Data != nil {
		l = len(m.Data)
		n += 1 + l + sovMerkledag(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovMerkledag(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMerkledag(x uint64) (n int) {
	return sovMerkledag(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func NewPopulatedPBLink(r randyMerkledag, easy bool) *PBLink {
	this := &PBLink{}
	if r.Intn(10) != 0 {
		v1 := r.Intn(100)
		this.Hash = make([]byte, v1)
		for i := 0; i < v1; i++ {
			this.Hash[i] = byte(r.Intn(256))
		}
	}
	if r.Intn(10) != 0 {
		v2 := randStringMerkledag(r)
		this.Name = &v2
	}
	if r.Intn(10) != 0 {
		v3 := uint64(r.Uint32())
		this.Tsize = &v3
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedMerkledag(r, 4)
	}
	return this
}

func NewPopulatedPBNode(r randyMerkledag, easy bool) *PBNode {
	this := &PBNode{}
	if r.Intn(10) != 0 {
		v4 := r.Intn(10)
		this.Links = make([]*PBLink, v4)
		for i := 0; i < v4; i++ {
			this.Links[i] = NewPopulatedPBLink(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		v5 := r.Intn(100)
		this.Data = make([]byte, v5)
		for i := 0; i < v5; i++ {
			this.Data[i] = byte(r.Intn(256))
		}
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedMerkledag(r, 3)
	}
	return this
}

type randyMerkledag interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneMerkledag(r randyMerkledag) rune {
	res := rune(r.Uint32() % 1112064)
	if 55296 <= res {
		res += 2047
	}
	return res
}
func randStringMerkledag(r randyMerkledag) string {
	v6 := r.Intn(100)
	tmps := make([]rune, v6)
	for i := 0; i < v6; i++ {
		tmps[i] = randUTF8RuneMerkledag(r)
	}
	return string(tmps)
}
func randUnrecognizedMerkledag(r randyMerkledag, maxFieldNumber int) (data []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		data = randFieldMerkledag(data, r, fieldNumber, wire)
	}
	return data
}
func randFieldMerkledag(data []byte, r randyMerkledag, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		data = encodeVarintPopulateMerkledag(data, uint64(key))
		v7 := r.Int63()
		if r.Intn(2) == 0 {
			v7 *= -1
		}
		data = encodeVarintPopulateMerkledag(data, uint64(v7))
	case 1:
		data = encodeVarintPopulateMerkledag(data, uint64(key))
		data = append(data, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		data = encodeVarintPopulateMerkledag(data, uint64(key))
		ll := r.Intn(100)
		data = encodeVarintPopulateMerkledag(data, uint64(ll))
		for j := 0; j < ll; j++ {
			data = append(data, byte(r.Intn(256)))
		}
	default:
		data = encodeVarintPopulateMerkledag(data, uint64(key))
		data = append(data, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return data
}
func encodeVarintPopulateMerkledag(data []byte, v uint64) []byte {
	for v >= 1<<7 {
		data = append(data, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	data = append(data, uint8(v))
	return data
}
func (m *PBLink) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *PBLink) MarshalTo(data []byte) (n int, err error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Hash != nil {
		data[i] = 0xa
		i++
		i = encodeVarintMerkledag(data, i, uint64(len(m.Hash)))
		i += copy(data[i:], m.Hash)
	}
	if m.Name != nil {
		data[i] = 0x12
		i++
		i = encodeVarintMerkledag(data, i, uint64(len(*m.Name)))
		i += copy(data[i:], *m.Name)
	}
	if m.Tsize != nil {
		data[i] = 0x18
		i++
		i = encodeVarintMerkledag(data, i, uint64(*m.Tsize))
	}
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}
func (m *PBNode) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *PBNode) MarshalTo(data []byte) (n int, err error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Links) > 0 {
		for _, msg := range m.Links {
			data[i] = 0x12
			i++
			i = encodeVarintMerkledag(data, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(data[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if m.Data != nil {
		data[i] = 0xa
		i++
		i = encodeVarintMerkledag(data, i, uint64(len(m.Data)))
		i += copy(data[i:], m.Data)
	}
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}
func encodeFixed64Merkledag(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Merkledag(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintMerkledag(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (this *PBLink) GoString() string {
	if this == nil {
		return "nil"
	}
	s := strings1.Join([]string{`&merkledag_pb.PBLink{` + `Hash:` + valueToGoStringMerkledag(this.Hash, "byte"), `Name:` + valueToGoStringMerkledag(this.Name, "string"), `Tsize:` + valueToGoStringMerkledag(this.Tsize, "uint64"), `XXX_unrecognized:` + fmt2.Sprintf("%#v", this.XXX_unrecognized) + `}`}, ", ")
	return s
}
func (this *PBNode) GoString() string {
	if this == nil {
		return "nil"
	}
	s := strings1.Join([]string{`&merkledag_pb.PBNode{` + `Links:` + fmt2.Sprintf("%#v", this.Links), `Data:` + valueToGoStringMerkledag(this.Data, "byte"), `XXX_unrecognized:` + fmt2.Sprintf("%#v", this.XXX_unrecognized) + `}`}, ", ")
	return s
}
func valueToGoStringMerkledag(v interface{}, typ string) string {
	rv := reflect1.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect1.Indirect(rv).Interface()
	return fmt2.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func extensionToGoStringMerkledag(e map[int32]code_google_com_p_gogoprotobuf_proto1.Extension) string {
	if e == nil {
		return "nil"
	}
	s := "map[int32]proto.Extension{"
	keys := make([]int, 0, len(e))
	for k := range e {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	ss := []string{}
	for _, k := range keys {
		ss = append(ss, strconv.Itoa(k)+": "+e[int32(k)].GoString())
	}
	s += strings1.Join(ss, ",") + "}"
	return s
}
func (this *PBLink) VerboseEqual(that interface{}) error {
	if that == nil {
		if this == nil {
			return nil
		}
		return fmt3.Errorf("that == nil && this != nil")
	}

	that1, ok := that.(*PBLink)
	if !ok {
		return fmt3.Errorf("that is not of type *PBLink")
	}
	if that1 == nil {
		if this == nil {
			return nil
		}
		return fmt3.Errorf("that is type *PBLink but is nil && this != nil")
	} else if this == nil {
		return fmt3.Errorf("that is type *PBLinkbut is not nil && this == nil")
	}
	if !bytes.Equal(this.Hash, that1.Hash) {
		return fmt3.Errorf("Hash this(%v) Not Equal that(%v)", this.Hash, that1.Hash)
	}
	if this.Name != nil && that1.Name != nil {
		if *this.Name != *that1.Name {
			return fmt3.Errorf("Name this(%v) Not Equal that(%v)", *this.Name, *that1.Name)
		}
	} else if this.Name != nil {
		return fmt3.Errorf("this.Name == nil && that.Name != nil")
	} else if that1.Name != nil {
		return fmt3.Errorf("Name this(%v) Not Equal that(%v)", this.Name, that1.Name)
	}
	if this.Tsize != nil && that1.Tsize != nil {
		if *this.Tsize != *that1.Tsize {
			return fmt3.Errorf("Tsize this(%v) Not Equal that(%v)", *this.Tsize, *that1.Tsize)
		}
	} else if this.Tsize != nil {
		return fmt3.Errorf("this.Tsize == nil && that.Tsize != nil")
	} else if that1.Tsize != nil {
		return fmt3.Errorf("Tsize this(%v) Not Equal that(%v)", this.Tsize, that1.Tsize)
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return fmt3.Errorf("XXX_unrecognized this(%v) Not Equal that(%v)", this.XXX_unrecognized, that1.XXX_unrecognized)
	}
	return nil
}
func (this *PBLink) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*PBLink)
	if !ok {
		return false
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.Hash, that1.Hash) {
		return false
	}
	if this.Name != nil && that1.Name != nil {
		if *this.Name != *that1.Name {
			return false
		}
	} else if this.Name != nil {
		return false
	} else if that1.Name != nil {
		return false
	}
	if this.Tsize != nil && that1.Tsize != nil {
		if *this.Tsize != *that1.Tsize {
			return false
		}
	} else if this.Tsize != nil {
		return false
	} else if that1.Tsize != nil {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *PBNode) VerboseEqual(that interface{}) error {
	if that == nil {
		if this == nil {
			return nil
		}
		return fmt3.Errorf("that == nil && this != nil")
	}

	that1, ok := that.(*PBNode)
	if !ok {
		return fmt3.Errorf("that is not of type *PBNode")
	}
	if that1 == nil {
		if this == nil {
			return nil
		}
		return fmt3.Errorf("that is type *PBNode but is nil && this != nil")
	} else if this == nil {
		return fmt3.Errorf("that is type *PBNodebut is not nil && this == nil")
	}
	if len(this.Links) != len(that1.Links) {
		return fmt3.Errorf("Links this(%v) Not Equal that(%v)", len(this.Links), len(that1.Links))
	}
	for i := range this.Links {
		if !this.Links[i].Equal(that1.Links[i]) {
			return fmt3.Errorf("Links this[%v](%v) Not Equal that[%v](%v)", i, this.Links[i], i, that1.Links[i])
		}
	}
	if !bytes.Equal(this.Data, that1.Data) {
		return fmt3.Errorf("Data this(%v) Not Equal that(%v)", this.Data, that1.Data)
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return fmt3.Errorf("XXX_unrecognized this(%v) Not Equal that(%v)", this.XXX_unrecognized, that1.XXX_unrecognized)
	}
	return nil
}
func (this *PBNode) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*PBNode)
	if !ok {
		return false
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if len(this.Links) != len(that1.Links) {
		return false
	}
	for i := range this.Links {
		if !this.Links[i].Equal(that1.Links[i]) {
			return false
		}
	}
	if !bytes.Equal(this.Data, that1.Data) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
