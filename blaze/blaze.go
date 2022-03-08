package blaze

import (
	"bufio"
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

type ReadCounter struct {
	Count int
}

// UInt16 reads an uint16 from the provided packet buffer using the
// big endian byte order
func (b *PacketBuff) UInt16(c *ReadCounter) uint16 {
	var out uint16
	_ = binary.Read(b, binary.BigEndian, &out)
	c.Count += 2
	return out
}

// UInt32 reads an uint32 from the provided packet buffer using the
// big endian byte order
func (b *PacketBuff) UInt32(c *ReadCounter) uint32 {
	var out uint32
	_ = binary.Read(b, binary.BigEndian, &out)
	c.Count += 4
	return out
}

// Float64 reads a float64 from the provided packet buffer using the
// big endian byte order
func (b *PacketBuff) Float64(c *ReadCounter) float64 {
	var out float64
	_ = binary.Read(b, binary.BigEndian, &out)
	c.Count += 8
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
func (b *PacketBuff) ReadVarInt(c *ReadCounter) uint64 {
	var x uint64
	var s uint
	for i := 0; i < 10; i++ {
		b, err := b.ReadByte()
		c.Count++
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
	_ = binary.Write(b.Writer, binary.BigEndian, value)
}

// ReadString reads a string from the buffer
func (b *PacketBuff) ReadString(c *ReadCounter) string {
	l := b.ReadVarInt(c)
	buf := make([]byte, l)
	_, _ = io.ReadFull(b, buf)
	_, _ = b.ReadByte() // Strings end with a zero byte
	c.Count += len(buf) + 1
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

var defaultCounter = &ReadCounter{}

// ReadPacket reads a game packet from the provided packet reader
func (b *PacketBuff) ReadPacket(c *ReadCounter) *Packet {
	if c == nil {
		c = defaultCounter
	}
	packet := Packet{
		Length:    b.UInt16(c),
		Component: b.UInt16(c),
		Command:   b.UInt16(c),
		Error:     b.UInt16(c),
		QType:     b.UInt16(c),
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = b.UInt16(c)
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
	c.Count += len(bytes)
	packet.Content = bytes
	return &packet
}

// ReadPacketHeading reads a game packet from the provided packet reader.
// but only reads the heading portion of the packet skips over the packet
// contents.
func (b *PacketBuff) ReadPacketHeading(c *ReadCounter) *Packet {
	if c == nil {
		c = defaultCounter
	}
	packet := Packet{
		Length:    b.UInt16(c),
		Component: b.UInt16(c),
		Command:   b.UInt16(c),
		Error:     b.UInt16(c),
		QType:     b.UInt16(c),
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = b.UInt16(c)
	} else {
		packet.ExtLength = 0
	}
	// Calculate the total size with the extension length
	l := int32(packet.Length) + (int32(packet.ExtLength) << 16)
	bytes := make([]byte, l) // Create an empty byte array in place of the content
	packet.Content = bytes
	return &packet
}

func (b *PacketBuff) ReadAllPackets() *list.List {
	out := list.New()
	size := b.Reader.Size()
	c := &ReadCounter{}
	for c.Count < size {
		out.PushBack(b.ReadPacket(c))
	}
	return out
}
