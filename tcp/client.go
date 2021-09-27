package tcp

import (
	"sync"
	"sync/atomic"
	"time"

	"errors"

	"github.com/sirupsen/logrus"
	"jinv/kim"
	"jinv/kim/logger"
)

type ClientOptions struct {
	Heartbeat time.Duration // 登录超时
	ReadWait  time.Duration // 读超时
	WriteWait time.Duration // 写超时
}

type Client struct {
	id   string
	name string

	conn kim.Conn

	state   int32
	options ClientOptions

	once sync.Once
	sync.Mutex

	kim.Dialer
}

func NewClient(id, name string, opts ClientOptions) kim.Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = kim.DefaultWriteWait
	}
	if opts.ReadWait == 0 {
		opts.ReadWait = kim.DefaultReadWait
	}

	return &Client{
		id:      id,
		name:    name,
		options: opts,
	}
}

func (c *Client) Connect(addr string) error {
	// 这是一个CAS原子操作，对比并设置值，是并发安全的
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return errors.New("client has connected")
	}

	rawconn, err := c.DialAndHandshake(kim.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: kim.DefaultLoginWait,
	})

	if err != nil {
		// 拨号失败时，要记得恢复客户端状态为"未连接"
		atomic.CompareAndSwapInt32(&c.state, 1, 0)

		return err
	}

	if rawconn == nil {
		return errors.New("conn is nil")
	}

	// 这里与 websocket 协议不同，主要是由于 websocket 返回的数据类型本身就是 Frame，不需要再次封装处理
	c.conn = NewConn(rawconn)

	if c.options.Heartbeat > 0 {
		go func() {
			if err := c.heartbeatLoop(); err != nil {
				// 心跳包发送失败的情况下，记录日志
				logger.WithField("module", "tcp.client").Warn("heartbeatLoop stopped - ", err)
			}
		}()
	}

	return nil
}

func (c *Client) SetDialer(dialer kim.Dialer) {
	c.Dialer = dialer
}

func (c *Client) Close() {
	c.once.Do(func() {
		// 平滑关闭与服务端的连接（四次挥手？）
		// 在关闭客户端连接之前，先通知服务端，然后在进行关闭

		// c.conn.WriteFrame(kim.OpClose, nil) todo 为什么没有使用 kim.Conn中的WriteFrame
		WriteFrame(c.conn, kim.OpClose, nil)

		c.conn.Close()

		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

func (c *Client) Read() (kim.Frame, error) {
	// todo 这里为什么不直接判断 c.state 呢，发送数据的时候是直接判断的 c.state
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}

	// todo 如果客户端与服务端之间有建立心跳，则在读取消息时设置超时时间 why
	if c.options.Heartbeat > 0 {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait)); err != nil {
			return nil, err
		}
	}

	frame, err := c.conn.ReadFrame()
	if err != nil {
		return nil, err
	}

	// 如果服务器主动断开连接的话，会收到 OpClose，这点的处理非常重要
	if frame.GetOpCode() == kim.OpClose {
		return nil, errors.New("remote close the channel")
	}

	return frame, nil
}

func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return errors.New("connection is nil")
	}

	// todo 这里加锁的目的是控制多个 goroutine 同时发送数据的情况？Read 怎么没有呢
	c.Lock()
	defer c.Unlock()

	if c.options.WriteWait > 0 {
		if err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait)); err != nil {
			return err
		}
	}

	return c.conn.WriteFrame(kim.OpBinary, payload)
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) heartbeatLoop() error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		if err := c.ping(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) ping() error {
	logrus.WithField("module", "tcp.client").Tracef("%s send ping to server", c.id)

	// 发送心跳包的时候也要加写超时呀
	if c.options.WriteWait > 0 {
		if err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait)); err != nil {
			return err
		}
	}

	return c.conn.WriteFrame(kim.OpPing, nil)
}
