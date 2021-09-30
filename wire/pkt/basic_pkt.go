package pkt

import (
	"io"

	"jinv/kim/wire/endian"
)

const (
	CodePing = uint16(1)
	CodePong = uint16(2)
)

type BasicPkt struct {
	Code   uint16
	Length uint16
	Body   []byte
}

func (pkt *BasicPkt) Decode(r io.Reader) error {
	var err error

	if pkt.Code, err = endian.ReadUint16(r); err != nil {
		return err
	}

	if pkt.Length, err = endian.ReadUint16(r); err != nil {
		return err
	}

	if pkt.Length > 0 {
		if pkt.Body, err = endian.ReadFixedBytes(int(pkt.Length), r); err != nil {
			return err
		}
	}

	return nil
}

func (pkt *BasicPkt) Encode(w io.Writer) error {
	if err := endian.WriteUint16(w, pkt.Code); err != nil {
		return err
	}

	if err := endian.WriteUint16(w, pkt.Length); err != nil {
		return nil
	}

	if pkt.Length > 0 {
		// todo 这里直接写了，可以不用包装了
		if _, err := w.Write(pkt.Body); err != nil {
			return err
		}
	}

	return nil
}
