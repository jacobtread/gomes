package blaze

import (
	"container/list"
	"encoding/binary"
	. "github.com/jacobtread/gomes/types"
	"io"
	"log"
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
	VarIntListType
	PairType
	TripleType
	FloatType

	EmptyType TdfType = 0x7F
)

type Tdf interface {
	Write(buf *PacketBuff)
	GetHead() TdfImpl
}

type TdfImpl struct {
	Label string
	Tag   uint32
	Type  TdfType

	Tdf
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

func (t Int64Tdf) GetHead() TdfImpl {
	return t.TdfImpl
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

func (t StringTdf) GetHead() TdfImpl {
	return t.TdfImpl
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

func (t BlobTdf) GetHead() TdfImpl {
	return t.TdfImpl
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

func (t StructTdf) Write(buf *PacketBuff) {
	if t.Start2 {
		_ = buf.WriteByte(2)
	}
	values := t.Values
	for l := values.Front(); l != nil; l = l.Next() {
		WriteTdf(buf, l.Value.(Tdf))
	}
	_ = buf.WriteByte(0)
}

func (t StructTdf) GetHead() TdfImpl {
	return t.TdfImpl
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

func (t ListTdf) GetHead() TdfImpl {
	return t.TdfImpl
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

func (t PairListTdf) GetHead() TdfImpl {
	return t.TdfImpl
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

func (t UnionTdf) GetHead() TdfImpl {
	return t.TdfImpl
}

type VarIntListTdf struct {
	Count int32
	List  list.List // List of int64
	TdfImpl
}

func NewVarIntList(label string, count int32, list list.List) VarIntListTdf {
	return VarIntListTdf{
		Count:   count,
		List:    list,
		TdfImpl: NewTdf(label, VarIntListType),
	}
}

func (t VarIntListTdf) Write(buf *PacketBuff) {
	buf.WriteVarInt(int64(t.Count))
	if t.Count > 0 {
		for l := t.List.Front(); l != nil; l = l.Next() {
			buf.WriteVarInt(l.Value.(int64))
		}
	}
}

func (t VarIntListTdf) GetHead() TdfImpl {
	return t.TdfImpl
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

func (t PairTdf) GetHead() TdfImpl {
	return t.TdfImpl
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

func (t TripleTdf) GetHead() TdfImpl {
	return t.TdfImpl
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

func (t FloatTdf) Write(buf *PacketBuff) {
	buf.WriteNum(t.Value)
}

func (t FloatTdf) GetHead() TdfImpl {
	return t.TdfImpl
}

func WriteTdf[T Tdf](buf *PacketBuff, value T) {
	head := value.GetHead()
	_ = binary.Write(buf, binary.BigEndian, head.Tag)
	_ = buf.WriteByte(byte(head.Type))
	value.Write(buf)
}

func ReadTdf(buf *PacketBuff) Tdf {
	head := buf.UInt32()
	tag := head & 0xFFFFFF00
	t := TdfType(head & 0xFF)
	impl := TdfImpl{
		Tag:   tag,
		Label: TagToLabel(tag),
		Type:  t,
	}
	switch t {
	case IntType:
		return ReadIntTdf(impl, buf)
	case StringType:
		return ReadStringTdf(impl, buf)
	case BlobType:
		return ReadBlobTdf(impl, buf)
	case StructType:
	case ListType:
	case PairListType:
	case UnionType:
	case VarIntListType:
	case PairType:
	case TripleType:
	case FloatType:
	default:
		log.Printf("Dont know how to handle tdf with type '%d'", t)
		return nil
	}
	return nil
}

func ReadIntTdf(head TdfImpl, buf *PacketBuff) Int64Tdf {
	return Int64Tdf{
		Value:   int64(buf.ReadVarInt()),
		TdfImpl: head,
	}
}

func ReadStringTdf(head TdfImpl, buf *PacketBuff) StringTdf {
	return StringTdf{
		Value:   buf.ReadString(),
		TdfImpl: head,
	}
}

func ReadBlobTdf(head TdfImpl, buf *PacketBuff) BlobTdf {
	size := buf.ReadVarInt()
	data := make([]byte, size)
	_, _ = io.ReadFull(buf, data)
	return BlobTdf{
		Data:    data,
		TdfImpl: head,
	}
}

func ReadStructValues(buf *PacketBuff) (*list.List, bool) {
	out := list.New()
	start2 := false
	for {
		b, err := buf.ReadByte()
		if err != nil {
			break
		}
		if b != 2 {
			_ = buf.UnreadByte()
		} else {
			start2 = true
		}
		out.PushBack(ReadTdf(buf))
	}
	return out, start2
}
