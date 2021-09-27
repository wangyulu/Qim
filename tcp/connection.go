package tcp

import (
	"io"
	"net"

	"jinv/kim"
	"jinv/kim/wire/endian"
)

type Frame struct {
	OpCode  kim.OpCode
	Payload []byte
}

func (f *Frame) SetOpCode(code kim.OpCode) {
	f.OpCode = code
}

func (f *Frame) GetOpCode() kim.OpCode {
	return f.OpCode
}

func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

func (f *Frame) GetPayload() []byte {
	return f.Payload
}

type TcpConn struct {
	net.Conn
}

func NewConn(conn net.Conn) *TcpConn {
	return &TcpConn{
		Conn: conn,
	}
}

func (c *TcpConn) ReadFrame() (kim.Frame, error) {
	opcode, err := endian.ReadUint8(c.Conn)
	if err != nil {
		return nil, err
	}

	payload, err := endian.ReadBytes(c.Conn)
	if err != nil {
		return nil, err
	}

	return &Frame{
		OpCode:  kim.OpCode(opcode),
		Payload: payload,
	}, nil
}

func (c *TcpConn) WriteFrame(code kim.OpCode, payload []byte) error {
	return WriteFrame(c.Conn, code, payload)
}

func (c *TcpConn) Flush() error {
	return nil
}

func WriteFrame(w io.Writer, code kim.OpCode, payload []byte) error {
	err := endian.WriteUint8(w, uint8(code))
	if err != nil {
		return err
	}

	err = endian.WriteBytes(w, payload)
	if err != nil {
		return err
	}

	return nil
}
