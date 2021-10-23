package kim

import (
	"bytes"

	"errors"

	"jinv/kim/wire/endian"
)

type Location struct {
	ChannelId string // 网关中的 channelID
	GateId    string // 网关 ID
}

func (loc *Location) Bytes() []byte {
	if loc == nil {
		return []byte{}
	}

	buf := new(bytes.Buffer)

	_ = endian.WriteShortBytes(buf, []byte(loc.ChannelId))
	_ = endian.WriteShortBytes(buf, []byte(loc.GateId))

	return buf.Bytes()
}

func (loc *Location) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return errors.New("data is empty")
	}

	buf := bytes.NewBuffer(data)

	var err error

	loc.ChannelId, err = endian.ReadShortString(buf)
	if err != nil {
		return err
	}

	loc.GateId, err = endian.ReadShortString(buf)

	return err
}
