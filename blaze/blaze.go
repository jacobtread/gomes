package blaze

import (
	"bufio"
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
	*bufio.Writer
	*bufio.Reader
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
	out, _ := binary.ReadUvarint(b)
	return out
}

// WriteNum takes any number type and writes it to the packet
func (b *PacketBuff) WriteNum(value any) {
	_ = binary.Write(b.Writer, binary.BigEndian, value)
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
	if strings.HasSuffix(value, "\\0") {
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
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = b.UInt16()
	} else {
		packet.ExtLength = 0
	}
	// Calculate the total size with the extension length
	l := int32(packet.Length) + (int32(packet.ExtLength) << 16)
	bytes := make([]byte, l)        // Create an empty byte array for the content
	_, err := io.ReadFull(b, bytes) // Read all the content bytes
	if err != nil {
		return nil
	}
	packet.Content = bytes
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
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = b.UInt16()
	} else {
		packet.ExtLength = 0
	}
	// Calculate the total size with the extension length
	l := int32(packet.Length) + (int32(packet.ExtLength) << 16)
	bytes := make([]byte, l) // Create an empty byte array in place of the content
	packet.Content = bytes
	return &packet
}
