package pkt

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"jinv/kim/wire"
)

func Test_ReadPkt(t *testing.T) {
	seq := wire.Seq.Next()

	packet := New(wire.CommandLoginSignIn, WithSeq(seq), WithStatus(Status_Success))
	packet.WriteBody(&LoginReq{
		Token: "token1",
	})
	packet.AddMeta(&Meta{
		Key:   "test1",
		Value: "test1",
	}, &Meta{
		Key:   "test",
		Value: "test",
	})

	bufs := new(bytes.Buffer)
	_ = packet.Encode(bufs)

	t.Log(bufs.Bytes())

	packet = &LogicPkt{}
	err := packet.Decode(bufs)
	assert.Nil(t, err)

	assert.Equal(t, wire.CommandLoginSignIn, packet.Command)
	assert.Equal(t, Status_Success, packet.Status)

	assert.Equal(t, 2, len(packet.Meta))

	packet.DelMeta("test1")
	assert.Equal(t, 1, len(packet.Meta))
}

func Benchmark_Encode(b *testing.B) {
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

	for i := 0; i < b.N; i++ {
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
	}
}

func Test_CusEncode(t *testing.T) {
	pkg := struct {
		Source   uint32
		Sequence uint64
		Data     []byte
	}{
		Source:   10000000,
		Sequence: 2<<60 + 3,
		Data:     []byte("hello world"),
	}

	// 大端序
	endian := binary.BigEndian

	buf := make([]byte, 1024)

	i := 0
	endian.PutUint32(buf[i:i+4], pkg.Source)

	i += 4
	endian.PutUint64(buf[i:i+8], pkg.Sequence)

	i += 8
	// 由于 data 的长度不确定，必须先把长度写入 buf，这样在反序列化时就可以正确的解析出 data
	dataLen := len(pkg.Data)
	endian.PutUint32(buf[i:i+4], uint32(dataLen))

	i += 4
	copy(buf[i:i+dataLen], pkg.Data)

	i += dataLen

	fmt.Println(buf[0:i], i)
	fmt.Println("length ", i)
}

func Test_CusDecode(t *testing.T) {
	var pkg struct {
		Source   uint32
		Sequence uint64
		Data     []byte
	}

	recv := []byte{0, 152, 150, 128, 32, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 11, 104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100}

	endian := binary.BigEndian

	i := 0
	pkg.Source = endian.Uint32(recv[i : i+4])

	i += 4
	pkg.Sequence = endian.Uint64(recv[i : i+8])

	i += 8
	dataLen := endian.Uint32(recv[i : i+4])

	i += 4
	pkg.Data = make([]byte, dataLen)

	copy(pkg.Data, recv[i:i+int(dataLen)])

	fmt.Printf("%d, %d, %s \n", pkg.Source, pkg.Sequence, pkg.Data)
}

func Benchmark_Cus(t *testing.B) {
	pkg := struct {
		Source   uint32
		Sequence uint64
		Data     []byte
	}{
		Source:   10000000,
		Sequence: 2<<60 + 3,
		Data:     []byte("hello world"),
	}

	// 大端序
	endian := binary.BigEndian

	buf := make([]byte, 1024)

	for i := 0; i < t.N; i++ {
		i := 0
		endian.PutUint32(buf[i:i+4], pkg.Source)

		i += 4
		endian.PutUint64(buf[i:i+8], pkg.Sequence)

		i += 8
		// 由于 data 的长度不确定，必须先把长度写入 buf，这样在反序列化时就可以正确的解析出 data
		dataLen := len(pkg.Data)
		endian.PutUint32(buf[i:i+4], uint32(dataLen))

		i += 4
		copy(buf[i:i+dataLen], pkg.Data)
	}
}

func Test_ProtoEncode(t *testing.T) {
	p := DemoPkt{
		Source:   10000000,
		Sequence: 2<<60 + 3,
		Data:     []byte("hello world"),
	}
	bts, _ := proto.Marshal(&p)
	t.Log(bts)
	t.Log("length ", len(bts))
}

func Benchmark_Proto(t *testing.B) {
	for i := 0; i < t.N; i++ {
		p := DemoPkt{
			Source:   10000000,
			Sequence: 2<<60 + 3,
			Data:     []byte("hello world"),
		}

		_, _ = proto.Marshal(&p)
	}
}

func Test_JsonEncode(t *testing.T) {
	p := DemoPkt{
		Source:   10000000,
		Sequence: 2<<60 + 3,
		Data:     []byte("hello world"),
	}

	bts, _ := json.Marshal(&p)
	t.Log(bts)
}

func Benchmark_Json(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := DemoPkt{
			Source:   10000000,
			Sequence: 2<<60 + 3,
			Data:     []byte("hello world"),
		}

		_, _ = json.Marshal(&p)
	}
}
