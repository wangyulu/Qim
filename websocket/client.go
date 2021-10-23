package websocket

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"jinv/kim"
	"jinv/kim/logger"
)

type ClientOptions struct {
	Heartbeat time.Duration // 登录超时
	ReadWait  time.Duration // 读超时
	WriteWait time.Duration // 写超时
}

type Client struct {
	sync.Mutex
	kim.Dialer
	once    sync.Once
	id      string
	name    string
	conn    net.Conn
	state   int32
	options ClientOptions
	Meta    map[string]string
}

func NewClient(id, name string, opts ClientOptions) kim.Client {
	return NewClientWithProps(id, name, make(map[string]string), opts)
}

func NewClientWithProps(id, name string, meta map[string]string, opts ClientOptions) kim.Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = kim.DefaultWriteWait
	}
	if opts.ReadWait == 0 {
		opts.ReadWait = kim.DefaultReadWait
	}

	cli := &Client{
		id:      id,
		name:    name,
		options: opts,
		Meta:    meta,
	}
	return cli
}

func (c *Client) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}

	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return errors.New("client has connected")
	}

	// 1. 拨号及握手（也就是与服务端建立连接，及用户认证的操作）
	conn, err := c.Dialer.DialAndHandshake(kim.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: kim.DefaultLoginWait,
	})

	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}

	if conn == nil {
		return fmt.Errorf("conn is nil")
	}

	c.conn = conn

	// 2. 如果有设置心跳时间，则进行心跳检测
	if c.options.Heartbeat > 0 {
		go func() {
			err := c.hearbeatLoop(conn)
			if err != nil {
				logger.Error("heartbeatLoop stopped", err)
			}
		}()
	}

	return nil
}

func (c *Client) SetDialer(dialer kim.Dialer) {
	c.Dialer = dialer
}

func (c *Client) Read() (kim.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}

	// todo 这里为何是有设置心跳检测时间时，才会设置读超时呢
	if c.options.Heartbeat > 0 {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait)); err != nil {
			return nil, err
		}
	}

	frame, err := ws.ReadFrame(c.conn)
	if err != nil {
		return nil, err
	}

	if frame.Header.OpCode == ws.OpClose {
		return nil, errors.New("remote side close the channel")
	}

	return &Frame{raw: frame}, nil
}

func (c *Client) Send(payload []byte) error {
	// 判断客户端是否已经启动
	if atomic.LoadInt32(&c.state) == 0 {
		return errors.New("connection is nil")
	}

	// todo 这里为什么也加锁了
	c.Lock()
	defer c.Unlock()

	// 设置写操作的超时时间
	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}

	// 客户端消息需要使用MASK
	return wsutil.WriteClientMessage(c.conn, ws.OpBinary, payload)
}

func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}

		// 平滑关闭与服务端的连接（四次挥手？）
		// 在关闭客户端连接之前，先通知服务端，然后在进行关闭
		_ = wsutil.WriteClientMessage(c.conn, ws.OpClose, nil)

		c.conn.Close()

		// 将客户端状态设置为未启动
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) hearbeatLoop(conn net.Conn) error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		if err := c.ping(conn); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) ping(conn net.Conn) error {
	// todo 这里为什么还需要加锁呢
	c.Lock()
	defer c.Unlock()

	// 设置写操作的超时时间
	err := conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return nil
	}

	logger.Tracef("%s send ping to server", c.id)

	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}

func (c *Client) ServiceID() string {
	return c.id
}

func (c *Client) ServiceName() string {
	return c.name
}

func (c *Client) GetMeta() map[string]string {
	return c.Meta
}
