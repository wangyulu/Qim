package endian

import (
	"encoding/binary"
	"testing"

	"github.com/magiconair/properties/assert"
)

func Test_ReadUint32(t *testing.T) {
	a := uint32(0x01020304)
	arr := make([]byte, 4)
	binary.BigEndian.PutUint32(arr, a)
	t.Log(arr)
	assert.Equal(t, byte(1), arr[0])

	binary.LittleEndian.PutUint32(arr, a)
	t.Log(arr)
	assert.Equal(t, byte(4), arr[0])
}

func Test_Encode(t *testing.T) {
	var pkt = struct {
		Source         uint16
		Destination    uint16
		Sequence       uint32
		Acknowledgment uint32
		Data           []byte
	}{
		Source:         2000,
		Destination:    90,
		Sequence:       100,
		Acknowledgment: 1,
		Data:           []byte("hello world"),
	}

	endian := binary.BigEndian

	buf := make([]byte, 1024)
	i := 0

	endian.PutUint16(buf[i:i+2], pkt.Source)
	i += 2

	endian.PutUint16(buf[i:i+2], pkt.Destination)
	i += 2

	endian.PutUint32(buf[i:i+4], pkt.Sequence)
	i += 4

	endian.PutUint32(buf[i:i+4], pkt.Acknowledgment)
	i += 4

	dataLen := len(pkt.Data)
	endian.PutUint32(buf[i:i+4], uint32(dataLen))
	i += 4

	copy(buf[i:i+dataLen], pkt.Data)
	i += dataLen

	t.Log(buf[0:i])
}

func Test_Decode(t *testing.T) {
	var pkt = struct {
		Source         uint16
		Destination    uint16
		Sequence       uint32
		Acknowledgment uint32
		Data           []byte
	}{}

	bts := []byte{7, 208, 0, 90, 0, 0, 0, 100, 0, 0, 0, 1, 0, 0, 0, 11, 104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100}

	endian := binary.BigEndian
	i := 0

	pkt.Source = endian.Uint16(bts[i : i+2])
	i += 2

	pkt.Destination = endian.Uint16(bts[i : i+2])
	i += 2

	pkt.Sequence = endian.Uint32(bts[i : i+4])
	i += 4

	pkt.Acknowledgment = endian.Uint32(bts[i : i+4])
	i += 4

	var dataLen uint32
	dataLen = endian.Uint32(bts[i : i+4])
	i += 4

	pkt.Data = make([]byte, dataLen)
	copy(pkt.Data, bts[i:i+int(dataLen)])

	t.Log(pkt)
	t.Logf("Src:%d Dest:%d Seq:%d Ack:%d Data:%s", pkt.Source, pkt.Destination, pkt.Sequence, pkt.Acknowledgment, pkt.Data)
}
