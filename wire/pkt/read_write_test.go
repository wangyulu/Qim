package pkt

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"jinv/kim/wire"
)

func Test_Marshal(t *testing.T) {
	basicPkt := &BasicPkt{
		Code: CodePong,
	}

	bts1 := Marshal(basicPkt)
	t.Log(bts1)
	assert.Equal(t, wire.MagicBasicPkt[1], bts1[1])
	assert.Equal(t, wire.MagicBasicPkt[3], bts1[3])

	logicPkt := New("login.sign.in")
	bts2 := Marshal(logicPkt)
	t.Log(bts2)

	assert.Equal(t, wire.MagicLogicPkt[1], bts2[1])
	assert.Equal(t, wire.MagicLogicPkt[3], bts2[3])
}

func Test_Read(t *testing.T) {
	bts := []byte{195, 17, 163, 101, 15, 0, 0, 0, 10, 13, 108, 111, 103, 105, 110, 46, 115, 105, 103, 110, 46, 105, 110, 0, 0, 0, 0}

	buf := bytes.NewBuffer(bts)

	logicPkt, err := Read(buf)
	assert.Nil(t, err)

	pkt, ok := logicPkt.(*LogicPkt)
	assert.Equal(t, true, ok)
	assert.Equal(t, "login.sign.in", pkt.Command)
}
