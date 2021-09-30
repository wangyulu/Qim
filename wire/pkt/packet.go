package pkt

import (
	"fmt"
	"io"
	"strings"

	"google.golang.org/protobuf/proto"
	"jinv/kim/wire/endian"
)

// LogicPkt
type LogicPkt struct {
	Header
	Body []byte `json:"body,omitempty"`
}

type HeaderOption func(header *Header)

func WithStatus(status Status) HeaderOption {
	return func(header *Header) {
		header.Status = status
	}
}

func WithSeq(seq uint32) HeaderOption {
	return func(header *Header) {
		header.Sequence = seq
	}
}

func WithChanncel(channelId string) HeaderOption {
	return func(header *Header) {
		header.ChannelId = channelId
	}
}

func WithDest(dest string) HeaderOption {
	return func(header *Header) {
		header.Dest = dest
	}
}

func New(command string, options ...HeaderOption) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Command = command

	for _, option := range options {
		option(&pkt.Header)
	}

	return pkt
}

func NewFrom(header *Header) *LogicPkt {
	pkt := &LogicPkt{}

	pkt.Header = Header{
		Command:   header.Command,
		Sequence:  header.Sequence,
		ChannelId: header.ChannelId,
		Status:    header.Status,
		Dest:      header.Dest,
	}

	return pkt
}

func (pkt *LogicPkt) Decode(r io.Reader) error {
	headerBytes, err := endian.ReadBytes(r)
	if err != nil {
		return err
	}

	if err := proto.Unmarshal(headerBytes, &pkt.Header); err != nil {
		return err
	}

	if pkt.Body, err = endian.ReadBytes(r); err != nil {
		return err
	}

	return nil
}

func (pkt *LogicPkt) Encode(w io.Writer) error {
	headerBytes, err := proto.Marshal(&pkt.Header)
	if err != nil {
		return err
	}

	if err = endian.WriteBytes(w, headerBytes); err != nil {
		return err
	}

	if err = endian.WriteBytes(w, pkt.Body); err != nil {
		return err
	}

	return nil
}

func (pkt *LogicPkt) ReadBody(val proto.Message) error {
	return proto.Unmarshal(pkt.Body, val)
}

func (pkt *LogicPkt) WriteBody(val proto.Message) *LogicPkt {
	if val == nil {
		return pkt
	}

	// todo 这里不用处理 err 的情况吗
	pkt.Body, _ = proto.Marshal(val)

	return pkt
}

func (pkt *LogicPkt) StringBody() string {
	return string(pkt.Body)
}

func (pkt *LogicPkt) String() string {
	return fmt.Sprintf("header: %v, body: %d bits", pkt.Header, len(pkt.Body))
}

// Header

func (header *Header) ServiceName() string {
	arr := strings.SplitN(header.Command, ".", 2)
	if len(arr) <= 1 {
		return "default"
	}

	return arr[0]
}
