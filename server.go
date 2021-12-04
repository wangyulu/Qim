package kim

import (
	"context"
	"net"
	"time"
)

const (
	DefaultReadWait  = time.Minute * 3
	DefaultWriteWait = time.Second * 10
	DefaultLoginWait = time.Second * 10
	DefaultHeartbeat = time.Second * 55
)

// 定义了基础服务的抽象接口
type Service interface {
	ServiceID() string
	ServiceName() string
	GetMeta() map[string]string
}

type ServiceRegistration interface {
	Service
	PublicAddress() string
	PublicPort() int
	DialURL() string
	GetTags() []string
	GetProtocol() string
	GetNamespace() string
	String() string
}

type Server interface {
	ServiceRegistration

	SetAcceptor(Acceptor)
	SetMessageListener(MessageListener)
	SetStateListener(StateListener)

	SetReadWait(duration time.Duration)
	SetChannelMap(ChannelMap)

	Start() error
	Shutdown(context.Context) error

	// Push 消息到指定的 Channel 中
	// channelID
	Push(string, []byte) error
}

type Acceptor interface {
	Accept(Conn, time.Duration) (string, Meta, error)
}

type StateListener interface {
	Disconnect(Agent) error
}

type MessageListener interface {
	Receive(Agent, []byte)
}

type Meta map[string]string

// 发送方
type Agent interface {
	ID() string
	Push([]byte) error
	GetMeta() Meta
}

type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}

type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}

type Channel interface {
	Conn
	Agent
	Close() error
	ReadLoop(lst MessageListener) error
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
}

type Client interface {
	Service

	ID() string
	Name() string

	Connect(string) error
	SetDialer(Dialer)
	Send([]byte) error
	Read() (Frame, error)
	Close()
}

type Dialer interface {
	DialAndHandshake(DialerContext) (net.Conn, error)
}

type DialerContext struct {
	Id      string
	Name    string
	Address string
	Timeout time.Duration
}
