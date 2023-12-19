package ws

import (
	"testing"
)

func TestWriteCreatesCorrectFrames(t *testing.T) {
	r := &WsUpgradeResult{}
	// r.bufrw =
	f := Frame{
		IsFinal:    false,
		Opcode:     0,
		Reserved:   0,
		IsMasked:   false,
		Length:     0,
		Payload:    []byte{},
		MaskingKey: []byte{},
	}
	r.Write(f)
}
