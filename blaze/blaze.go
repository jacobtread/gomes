package blaze

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"io"
	"net"
	"strings"
)

type Connection struct {
	PacketBuff
	net.Conn
}

type PacketBuff struct {
	*bytes.Buffer
}

type Packet struct {
	Length    uint16
	Component uint16
	Command   uint16
	Error     uint16
	QType     uint16
	Id        uint16
	ExtLength uint16
	Content   []byte
}

// UInt16 reads an uint16 from the provided packet buffer using the
// big endian byte order
func (b *PacketBuff) UInt16() uint16 {
	var out uint16
	_ = binary.Read(b, binary.BigEndian, &out)
	return out
}

// UInt32 reads an uint32 from the provided packet buffer using the
// big endian byte order
func (b *PacketBuff) UInt32() uint32 {
	var out uint32
	_ = binary.Read(b, binary.BigEndian, &out)
	return out
}

// Float64 reads a float64 from the provided packet buffer using the
// big endian byte order
func (b *PacketBuff) Float64() float64 {
	var out float64
	_ = binary.Read(b, binary.BigEndian, &out)
	return out
}

// WriteVarInt writes a var int to the packet buffer
func (b *PacketBuff) WriteVarInt(value int64) {
	ux := uint64(value) << 1
	if value < 0 {
		ux = ^ux
	}
	i := 0
	for ux >= 0x80 {
		_ = b.WriteByte(byte(ux) | 0x80)
		ux >>= 7
		i++
	}
	_ = b.WriteByte(byte(ux))
}

// ReadVarInt reads a var int from the packet buffer
func (b *PacketBuff) ReadVarInt() uint64 {
	var x uint64
	var s uint
	for i := 0; i < 10; i++ {
		b, err := b.ReadByte()
		if err != nil {
			return x
		}
		if b < 0x80 {
			if i == 9 && b > 1 {
				return x
			}
			return x | uint64(b)<<s
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return x
}

// WriteNum takes any number type and writes it to the packet
func (b *PacketBuff) WriteNum(value any) {
	_ = binary.Write(b, binary.BigEndian, value)
}

// ReadString reads a string from the buffer
func (b *PacketBuff) ReadString() string {
	l := b.ReadVarInt()
	buf := make([]byte, l)
	_, _ = io.ReadFull(b, buf)
	_, _ = b.ReadByte() // Strings end with a zero byte
	return string(buf)
}

// WriteString writes a string to the buffer
func (b *PacketBuff) WriteString(value string) {
	var l int
	if strings.HasSuffix(value, "\x00") {
		l = len(value)
	} else {
		l = len(value) + 1
	}
	b.WriteVarInt(int64(l))
	_, _ = b.Write([]byte(value))
	_ = b.WriteByte(0)
}

// ReadPacket reads a game packet from the provided packet reader
func (b *PacketBuff) ReadPacket() *Packet {
	packet := Packet{
		Length:    b.UInt16(),
		Component: b.UInt16(),
		Command:   b.UInt16(),
		Error:     b.UInt16(),
		QType:     b.UInt16(),
		Id:        b.UInt16(),
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = b.UInt16()
	} else {
		packet.ExtLength = 0
	}
	// Calculate the total size with the extension length
	l := int32(packet.Length) + (int32(packet.ExtLength) << 16)
	by := make([]byte, l)        // Create an empty byte array for the content
	_, err := io.ReadFull(b, by) // Read all the content bytes
	if err != nil {
		return nil
	}
	packet.Content = by
	return &packet
}

// ReadPacketHeading reads a game packet from the provided packet reader.
// but only reads the heading portion of the packet skips over the packet
// contents.
func (b *PacketBuff) ReadPacketHeading() *Packet {
	packet := Packet{
		Length:    b.UInt16(),
		Component: b.UInt16(),
		Command:   b.UInt16(),
		Error:     b.UInt16(),
		QType:     b.UInt16(),
		Id:        b.UInt16(),
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = b.UInt16()
	} else {
		packet.ExtLength = 0
	}
	// Calculate the total size with the extension length
	l := int32(packet.Length) + (int32(packet.ExtLength) << 16)
	by := make([]byte, l) // Create an empty byte array in place of the content
	packet.Content = by
	return &packet
}

func (b *PacketBuff) ReadAllPackets() *list.List {
	out := list.New()
	for b.Len() > 0 {
		out.PushBack(b.ReadPacket())
	}
	return out
}

func (b *PacketBuff) EncodePacket(comp uint16, cmd uint16, err uint16, qType uint16, id uint16, content list.List) []byte {
	buf := &PacketBuff{Buffer: &bytes.Buffer{}}
	contentBuff := &PacketBuff{Buffer: &bytes.Buffer{}}
	for l := content.Front(); l != nil; l = l.Next() {
		WriteTdf(contentBuff, l.Value.(Tdf))
	}
	c := contentBuff.Bytes()
	l := len(c)

	_ = buf.WriteByte(byte((l & 0xFFFF) >> 8))
	_ = buf.WriteByte(byte(l & 0xFF))
	_ = buf.WriteByte(0)
	_ = binary.Write(buf, binary.BigEndian, comp)
	_ = binary.Write(buf, binary.BigEndian, cmd)
	_ = binary.Write(buf, binary.BigEndian, err)

	buf.WriteByte(byte(qType >> 8))
	if l > 0xFFFF {
		buf.WriteByte(0x10)
	} else {
		buf.WriteByte(0x00)
	}

	_ = binary.Write(buf, binary.BigEndian, id)

	if l > 0xFFFF {
		buf.WriteByte(byte((l & 0xFF000000) >> 24))
		buf.WriteByte(byte((l & 0x00FF0000) >> 16))
	}

	_, _ = buf.Write(c)
	return buf.Bytes()
}

func (b *PacketBuff) EncodePacketRaw(packet Packet) []byte {
	buf := &PacketBuff{Buffer: &bytes.Buffer{}}
	_ = buf.WriteByte(byte(packet.Length >> 8))
	_ = buf.WriteByte(byte(packet.Length & 0xFF))
	_ = buf.WriteByte(0)
	_ = binary.Write(buf, binary.BigEndian, packet.Component)
	_ = binary.Write(buf, binary.BigEndian, packet.Command)
	_ = binary.Write(buf, binary.BigEndian, packet.Error)
	_ = binary.Write(buf, binary.BigEndian, packet.QType)
	_ = binary.Write(buf, binary.BigEndian, packet.Id)
	if (packet.QType & 0x10) != 0 {
		buf.WriteByte(byte(packet.ExtLength >> 8))
		buf.WriteByte(byte(packet.ExtLength & 0xFF))
	}
	_, _ = buf.Write(packet.Content)
	return buf.Bytes()
}
