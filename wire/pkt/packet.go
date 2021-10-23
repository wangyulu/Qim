package pkt

import (
	"fmt"
	"io"
	"strconv"
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

func WithChannel(channelId string) HeaderOption {
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
	p := &LogicPkt{}
	p.Command = command

	for _, option := range options {
		option(&p.Header)
	}

	return p
}

func NewFrom(header *Header) *LogicPkt {
	p := &LogicPkt{}

	p.Header = Header{
		Command:   header.Command,
		Sequence:  header.Sequence,
		ChannelId: header.ChannelId,
		Status:    header.Status,
		Dest:      header.Dest,
	}

	return p
}

func (p *LogicPkt) Decode(r io.Reader) error {
	headerBytes, err := endian.ReadBytes(r)
	if err != nil {
		return err
	}

	if err := proto.Unmarshal(headerBytes, &p.Header); err != nil {
		return err
	}

	if p.Body, err = endian.ReadBytes(r); err != nil {
		return err
	}

	return nil
}

func (p *LogicPkt) Encode(w io.Writer) error {
	headerBytes, err := proto.Marshal(&p.Header)
	if err != nil {
		return err
	}

	if err = endian.WriteBytes(w, headerBytes); err != nil {
		return err
	}

	if err = endian.WriteBytes(w, p.Body); err != nil {
		return err
	}

	return nil
}

func (p *LogicPkt) ReadBody(val proto.Message) error {
	return proto.Unmarshal(p.Body, val)
}

func (p *LogicPkt) WriteBody(val proto.Message) *LogicPkt {
	if val == nil {
		return p
	}

	// todo 这里不用处理 err 的情况吗
	p.Body, _ = proto.Marshal(val)

	return p
}

func (p *LogicPkt) StringBody() string {
	return string(p.Body)
}

func (p *LogicPkt) String() string {
	return fmt.Sprintf("header: %v, body: %d bits", p.Header, len(p.Body))
}

func (p *LogicPkt) AddMeta(m ...*Meta) {
	p.Meta = append(p.Meta, m...)
}

func (p *LogicPkt) AddStringMeta(key, value string) {
	p.AddMeta(&Meta{
		Key:   key,
		Value: value,
		Type:  MetaType_string,
	})
}

func (p *LogicPkt) GetMeta(key string) (interface{}, bool) {
	for _, m := range p.Meta {
		if m.Key == key {
			switch m.Type {
			case MetaType_int:
				v, _ := strconv.Atoi(m.Value)
				return v, true
			case MetaType_float:
				v, _ := strconv.ParseFloat(m.Value, 64)
				return v, true
			}
			return m.Value, true
		}
	}

	return nil, false
}

func (p *LogicPkt) DelMeta(key string) {
	for i, m := range p.Meta {
		if m.Key == key {
			length := len(p.Meta)

			if i < length-1 {
				copy(p.Meta[i:], p.Meta[i+1:])
			}

			p.Meta = p.Meta[:length-1]
		}
	}
}

// Header
func (header *Header) ServiceName() string {
	arr := strings.SplitN(header.Command, ".", 2)
	if len(arr) <= 1 {
		return "default"
	}

	return arr[0]
}
