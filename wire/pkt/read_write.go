package pkt

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"jinv/kim/wire"
)

type Packet interface {
	Decode(io.Reader) error
	Encode(io.Writer) error
}

func MustReadLogicPkt(r io.Reader) (*LogicPkt, error) {
	val, err := Read(r)

	if err != nil {
		return nil, err
	}

	if pkt, ok := val.(*LogicPkt); ok {
		return pkt, nil
	}

	return nil, fmt.Errorf("packet is not a logic packet")
}

func MustReadBasePkt(r io.Reader) (*BasicPkt, error) {
	val, err := Read(r)

	if err != nil {
		return nil, err
	}

	if pkt, ok := val.(*BasicPkt); ok {
		return pkt, nil
	}

	return nil, fmt.Errorf("packet is not a basic packet")
}

// 解包，返回基本协议包/逻辑协议包
func Read(r io.Reader) (interface{}, error) {
	magic := wire.Magic{}

	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, err
	}

	switch magic {
	case wire.MagicBasicPkt:
		pkt := new(BasicPkt)
		if err := pkt.Decode(r); err != nil {
			return nil, err
		}

		return pkt, nil
	case wire.MagicLogicPkt:
		pkt := new(LogicPkt)
		if err := pkt.Decode(r); err != nil {
			return nil, err
		}

		return pkt, nil
	default:
		return nil, fmt.Errorf("magic code %s is incorrect", magic)
	}
}

func Marshal(pkt Packet) []byte {
	buf := new(bytes.Buffer)

	kind := reflect.TypeOf(pkt).Elem()

	if kind.AssignableTo(reflect.TypeOf(LogicPkt{})) {
		_, _ = buf.Write(wire.MagicLogicPkt[:])
	} else if kind.AssignableTo(reflect.TypeOf(BasicPkt{})) {
		_, _ = buf.Write(wire.MagicBasicPkt[:])
	}

	// todo 这里没有处理 error 的情况
	_ = pkt.Encode(buf)

	return buf.Bytes()
}
