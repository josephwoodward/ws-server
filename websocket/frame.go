package ws

import (
	"bytes"
	"encoding/binary"
)

type WsOpCode int

const (
	// From https://tools.ietf.org/html/rfc6455#section-5.2
	WsTextMessage   = WsOpCode(1)
	WsBinaryMessage = WsOpCode(2)
	WsCloseMessage  = WsOpCode(8)
	WsPingMessage   = WsOpCode(9)
	WsPongMessage   = WsOpCode(10)
)

// https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
type Frame struct {
	IsFragment bool // if the
	Opcode2    byte
	Opcode     WsOpCode
	Reserved   byte
	IsMasked   bool
	Length     uint64
	Payload    []byte
	MaskingKey []byte
}

// Get the Pong frame
func (f Frame) Pong() Frame {
	f.Opcode2 = 10
	return f
}

// Get Text Payload
func (f Frame) Text() string {
	return string(f.Payload)
}

// IsControl checks if the frame is a control frame identified by opcodes where the most significant bit of the opcode is 1
func (f *Frame) IsControl() bool {
	return f.Opcode2&0x08 == 0x08
}

func (f *Frame) HasReservedOpCode() bool {
	return f.Opcode2 > 10 || (f.Opcode2 >= 3 && f.Opcode2 <= 7)
}

func (f *Frame) CloseCode() uint16 {
	var code uint16
	binary.Read(bytes.NewReader(f.Payload), binary.BigEndian, &code)
	return code
}
