package blaze

import (
	"container/list"
	"encoding/binary"
	. "github.com/jacobtread/gomes/types"
)

type TdfType byte

const (
	IntType TdfType = iota
	StringType
	BlobType
	StructType
	ListType
	PairListType
	UnionType
	IntListType
	PairType
	TripleType
	FloatType

	EmptyType TdfType = 0x7F
)

type Tdf interface {
	GetType() byte
	GetTag() uint32
	GetLabel() string

	Write(buf *PacketBuff)
}

type TdfImpl struct {
	Label string
	Tag   uint32
	Type  TdfType

	Tdf
}

func (t TdfImpl) GetType() TdfType {
	return t.Type
}
func (t TdfImpl) GetTag() uint32 {
	return t.Tag
}

func (t TdfImpl) GetLabel() string {
	return t.Label
}

func NewTdf(label string, t TdfType) TdfImpl {
	return TdfImpl{
		Label: label,
		Type:  t,
		Tag:   LabelToTag(label),
	}
}

func LabelToTag(label string) uint32 {
	res := make([]byte, 3)
	for len(label) < 4 {
		label += "\x00"
	}
	if len(label) > 4 {
		label = label[0:4]
	}
	buff := []byte(label)
	res[0] |= (buff[0] & 0x40) << 1
	res[0] |= (buff[0] & 0x10) << 2
	res[0] |= (buff[0] & 0x0F) << 2
	res[0] |= (buff[1] & 0x40) >> 5
	res[0] |= (buff[1] & 0x10) >> 4

	res[1] |= (buff[1] & 0x0F) << 4
	res[1] |= (buff[2] & 0x40) >> 3
	res[1] |= (buff[2] & 0x10) >> 2
	res[1] |= (buff[2] & 0x0C) >> 2

	res[2] |= (buff[2] & 0x03) << 6
	res[2] |= (buff[3] & 0x40) >> 1
	res[2] |= buff[3] & 0x1F

	return binary.BigEndian.Uint32(res)
}

func TagToLabel(tag uint32) string {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, tag)

	res := make([]byte, 4)

	res[0] |= (buff[3] & 0x40) << 1
	res[0] |= (buff[3] & 0x10) << 2
	res[0] |= (buff[3] & 0x0F) << 2
	res[0] |= (buff[2] & 0x40) >> 5
	res[0] |= (buff[2] & 0x10) >> 4

	res[1] |= (buff[2] & 0x0F) << 4
	res[1] |= (buff[1] & 0x40) >> 3
	res[1] |= (buff[1] & 0x10) >> 2
	res[1] |= (buff[1] & 0x0C) >> 2

	res[2] |= (buff[1] & 0x03) << 6
	res[2] |= (buff[0] & 0x40) >> 1
	res[2] |= buff[0] & 0x1F

	for i := 0; i < 4; i++ {
		if res[i] == 0 {
			res[i] = 0x20
		}
	}
	return string(res)
}

type Int64Tdf struct {
	Value int64
	TdfImpl
}

func NewInt64(label string, value int64) Int64Tdf {
	return Int64Tdf{
		Value:   value,
		TdfImpl: NewTdf(label, IntType),
	}
}

func (t Int64Tdf) Write(buf *PacketBuff) {
	buf.WriteVarInt(t.Value)
}

type StringTdf struct {
	Value string
	TdfImpl
}

func NewString(label string, value string) StringTdf {
	return StringTdf{
		Value:   value,
		TdfImpl: NewTdf(label, StringType),
	}
}

func (t StringTdf) Write(buf *PacketBuff) {
	buf.WriteString(t.Value)
}

type BlobTdf struct {
	Data []byte
	TdfImpl
}

func NewBlob(label string, data []byte) BlobTdf {
	return BlobTdf{
		Data:    data,
		TdfImpl: NewTdf(label, BlobType),
	}
}

func (t BlobTdf) Write(buf *PacketBuff) {
	data := t.Data
	buf.WriteVarInt(int64(len(data)))
	_, _ = buf.Write(data)
}

type StructTdf struct {
	Values list.List // List of TdfImpl values
	Start2 bool

	TdfImpl
}

func NewStruct(label string, values list.List) StructTdf {
	return StructTdf{
		Values:  values,
		TdfImpl: NewTdf(label, StructType),
		Start2:  false,
	}
}

func NewStruct2(label string, values list.List) StructTdf {
	return StructTdf{
		Values:  values,
		TdfImpl: NewTdf(label, StructType),
		Start2:  true,
	}
}

func (s StructTdf) Write(buf *PacketBuff) {
	if s.Start2 {
		_ = buf.WriteByte(2)
	}
	values := s.Values
	for l := values.Front(); l != nil; l = l.Next() {
		WriteTdf(buf, l.Value)
	}
	_ = buf.WriteByte(0)
}

type SubType = byte

const (
	IntList    SubType = 0
	StringList         = 1
	StructList         = 3
	TripleList         = 9
	FloatList          = 10
)

type ListTdf struct {
	SubType SubType
	Count   int32
	List    list.List

	TdfImpl
}

func NewList(label string, subtype SubType, count int32, list list.List) ListTdf {
	return ListTdf{
		SubType: subtype,
		Count:   count,
		List:    list,
		TdfImpl: NewTdf(label, ListType),
	}
}

func (l ListTdf) Write(buf *PacketBuff) {
	_ = buf.WriteByte(l.SubType)
	buf.WriteVarInt(int64(l.Count))
	for el := l.List.Front(); el != nil; el = el.Next() {
		switch l.SubType {
		case IntList:
			buf.WriteVarInt(el.Value.(int64))
		case StringList:
			buf.WriteString(el.Value.(string))
		case StructList:
			el.Value.(StructTdf).Write(buf)
		case TripleList:
			el.Value.(TripleTdf).Write(buf)
		}
	}
}

type PairListTdf struct {
	SubTypeA SubType
	SubTypeB SubType
	Count    int32

	ListA list.List
	ListB list.List

	TdfImpl
}

func NewPairList(label string, subtypeA SubType, subtypeB SubType, listA list.List, listB list.List, count int32) PairListTdf {
	return PairListTdf{
		SubTypeA: subtypeA,
		SubTypeB: subtypeB,
		ListA:    listA,
		ListB:    listB,
		Count:    count,
		TdfImpl:  NewTdf(label, PairListType),
	}
}

func (l PairListTdf) Write(buf *PacketBuff) {
	buf.WriteByte(l.SubTypeA)
	buf.WriteByte(l.SubTypeB)
	buf.WriteVarInt(int64(l.Count))

	listA := l.ListA
	listB := l.ListB
	for {
		a := listA.Front()
		b := listB.Front()

		if a == nil || b == nil {
			break
		}

		switch l.SubTypeA {
		case IntList:
			buf.WriteVarInt(a.Value.(int64))
		case StringList:
			buf.WriteString(a.Value.(string))
		case StructList:
			a.Value.(StructTdf).Write(buf)
		case FloatList:
			buf.WriteNum(a.Value.(float64))
		}

		switch l.SubTypeB {
		case IntList:
			buf.WriteVarInt(b.Value.(int64))
		case StringList:
			buf.WriteString(b.Value.(string))
		case StructList:
			b.Value.(StructTdf).Write(buf)
		case FloatList:
			buf.WriteNum(b.Value.(float64))
		}

		a.Next()
		b.Next()
	}
}

type UnionTdf struct {
	Type    TdfType
	Content Tdf
	TdfImpl
}

func NewUnion(label string, uType TdfType, value Tdf) UnionTdf {
	return UnionTdf{
		Type:    uType,
		Content: value,
		TdfImpl: NewTdf(label, UnionType),
	}
}

func (t UnionTdf) Write(buf *PacketBuff) {
	buf.WriteByte(byte(t.Type))
	if t.Type != EmptyType {
		WriteTdf(buf, t.Content)
	}
}

type IntListTdf struct {
	Count int32
	List  list.List // List of int64
	TdfImpl
}

func NewIntList(label string, count int32, list list.List) IntListTdf {
	return IntListTdf{
		Count:   count,
		List:    list,
		TdfImpl: NewTdf(label, IntListType),
	}
}

func (t IntListTdf) Write(buf *PacketBuff) {
	
}

type PairTdf struct {
	Pair
	TdfImpl
}

func NewPair(label string, value Pair) PairTdf {
	return PairTdf{
		Pair:    value,
		TdfImpl: NewTdf(label, PairType),
	}
}

func (p PairTdf) Write(buf *PacketBuff) {
	buf.WriteVarInt(p.A)
	buf.WriteVarInt(p.B)
}

type TripleTdf struct {
	Triple
	TdfImpl
}

func NewTriple(label string, value Triple) TripleTdf {
	return TripleTdf{
		Triple:  value,
		TdfImpl: NewTdf(label, TripleType),
	}
}

func (t TripleTdf) Write(buf *PacketBuff) {
	buf.WriteVarInt(t.A)
	buf.WriteVarInt(t.B)
	buf.WriteVarInt(t.C)
}

type FloatTdf struct {
	Value float64
	TdfImpl
}

func NewFloat(label string, value float64) FloatTdf {
	return FloatTdf{
		Value:   value,
		TdfImpl: NewTdf(label, FloatType),
	}
}

func WriteTdf[T Tdf](buf *PacketBuff, value T) {
	_ = binary.Write(buf, binary.BigEndian, value.GetTag())
	_ = buf.WriteByte(value.GetType())
	value.Write(buf)
}
