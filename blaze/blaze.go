package blaze

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
)

type Connection struct {
	PacketBuff
	net.Conn
}

type PacketBuff struct {
	*bufio.ReadWriter
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
func (r *PacketBuff) UInt16() uint16 {
	var out uint16
	_ = binary.Read(r, binary.BigEndian, &out)
	return out
}

// UInt32 reads an uint32 from the provided packet buffer using the
// big endian byte order
func (r *PacketBuff) UInt32() uint32 {
	var out uint32
	_ = binary.Read(r, binary.BigEndian, &out)
	return out
}

// Float64 reads a float64 from the provided packet buffer using the
// big endian byte order
func (r *PacketBuff) Float64() float64 {
	var out float64
	_ = binary.Read(r, binary.BigEndian, &out)
	return out
}

// ReadPacket reads a game packet from the provided packet reader
func (r *PacketBuff) ReadPacket() *Packet {
	packet := Packet{
		Length:    r.UInt16(),
		Component: r.UInt16(),
		Command:   r.UInt16(),
		Error:     r.UInt16(),
		QType:     r.UInt16(),
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = r.UInt16()
	} else {
		packet.ExtLength = 0
	}
	// Calculate the total size with the extension length
	l := int32(packet.Length) + (int32(packet.ExtLength) << 16)
	bytes := make([]byte, l)        // Create an empty byte array for the content
	_, err := io.ReadFull(r, bytes) // Read all the content bytes
	if err != nil {
		return nil
	}
	packet.Content = bytes
	return &packet
}

// ReadPacketHeading reads a game packet from the provided packet reader.
// but only reads the heading portion of the packet skips over the packet
// contents.
func (r *PacketBuff) ReadPacketHeading() *Packet {
	packet := Packet{
		Length:    r.UInt16(),
		Component: r.UInt16(),
		Command:   r.UInt16(),
		Error:     r.UInt16(),
		QType:     r.UInt16(),
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = r.UInt16()
	} else {
		packet.ExtLength = 0
	}
	// Calculate the total size with the extension length
	l := int32(packet.Length) + (int32(packet.ExtLength) << 16)
	bytes := make([]byte, l) // Create an empty byte array in place of the content
	packet.Content = bytes
	return &packet
}
