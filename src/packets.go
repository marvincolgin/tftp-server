package main

import (
	"bytes"
	"encoding/binary"
	"net"
)

// RFC: https://www.ietf.org/rfc/rfc1350.txt (Page 9)
const (
	ErrorNotDefined          uint16 = iota
	ErrorFileNotFound        uint16 = iota
	ErrorFileAccessViolation uint16 = iota
	ErrorDiskFull            uint16 = iota
	ErrorIllegalOp           uint16 = iota
	ErrorUnknownTID          uint16 = iota
	ErrorFileExists          uint16 = iota
	ErrorUnknownUser         uint16 = iota
)

// NewPacketError will create the struct PacketError
func NewPacketError(code uint16, msg string) PacketError {

	packet := PacketError{}
	packet.Code = code
	packet.Msg = msg

	return packet
}

// makePacketRequest will take a generic PacketRequest and build a PacketReqest from it
func makePacketRequest(buf []byte) PacketRequest {
	// I had some issues with casting/using a raw 'Packet' into a 'PacketRequest'
	// So, I'm just going to copy the data out (from .Serialize()) explicityly

	p := PacketRequest{}

	// Op is first 16 bits (2x 8-bit 'chars')
	p.Op = binary.BigEndian.Uint16(buf[:2])

	// First 0x00 character will be file-name string delimiter
	n := bytes.Index(buf[2:], []byte{0})
	p.Filename = string(buf[2 : n+2])

	// Mode is the last character before the final 0x00 terminater
	m := bytes.Index(buf[n+3:], []byte{0})
	p.Mode = string(buf[n+3 : n+3+m])

	return p
}

// makePacketData will create a data packet from params
func makePacketData(blockNum uint16, buf []byte, pos int, size int) PacketData {

	p := PacketData{}

	p.BlockNum = blockNum

	p.Data = make([]byte, size)
	copy(p.Data, buf[pos:pos+size])

	return p
}

// RawPacket is the raw-bytes received over wire, with the RemoteAddr saved
type RawPacket struct {
	Addr  *net.UDPAddr
	bytes []byte
}

func (packet RawPacket) getBytes() []byte {
	return packet.bytes
}

func (packet RawPacket) getAddr() *net.UDPAddr {
	return packet.Addr
}
