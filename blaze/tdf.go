package blaze

import (
	"container/list"
	"encoding/binary"
)

const (
	IntType = iota
	StringType
	BlobType
	StructType
	ListType
	DoubleListType
	UnionType
	IntListType
	DoubleType
	TripleType
	FloatType
)

type TdfWritable interface {
	Write(buf *PacketBuff)
}

type Tdf struct {
	Label string
	Tag   uint32
	Type  byte

	TdfWritable
}

type Pair struct {
	A int64
	B int64

	Tdf
}

type Triple struct {
	A int64
	B int64
	C int64

	Tdf
}

func NewTdf(label string, t byte) Tdf {
	return Tdf{
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

	Tdf
}

func (t Int64Tdf) Write(buf *PacketBuff) {
	buf.WriteVarInt(t.Value)
}

func NewIntTdf(label string, value int64) Int64Tdf {
	return Int64Tdf{
		Value: value,
		Tdf:   NewTdf(label, IntType),
	}
}

type FloatTdf struct {
	Value float64

	Tdf
}

func (t FloatTdf) Write(buf *PacketBuff) {

}

func NewFloatTdf(label string, value float64) FloatTdf {
	return FloatTdf{
		Value: value,
		Tdf:   NewTdf(label, FloatType),
	}
}

type StringTdf struct {
	Value string

	Tdf
}

func NewStringTdf(label string, value string) StringTdf {
	return StringTdf{
		Value: value,
		Tdf:   NewTdf(label, StringType),
	}
}

type StructTdf struct {
	Values list.List // List of Tdf values
	Start2 bool

	Tdf
}

func NewStructTdf(label string, values list.List) StructTdf {
	return StructTdf{
		Values: values,
		Tdf:    NewTdf(label, StructType),
		Start2: false,
	}
}

func NewStructTdf2(label string, values list.List) StructTdf {
	return StructTdf{
		Values: values,
		Tdf:    NewTdf(label, StructType),
		Start2: true,
	}
}

func WriteTdf[T Tdf](buf *PacketBuff, value T) {
	_ = binary.Write(buf, binary.BigEndian, value.Tag)
	_ = buf.WriteByte(value.Type)
	value.Write(buf)
}
