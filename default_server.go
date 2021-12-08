package kim

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gobwas/pool/pbufio"
	"github.com/gobwas/ws"
	"github.com/prometheus/common/log"
	"github.com/segmentio/ksuid"
	"jinv/kim/logger"
)

type Upgrader interface {
	Name() string
	Upgrade(rawconn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (Conn, error)
}

type ServerOptions struct {
	LoginWait time.Duration // 登录超时
	ReadWait  time.Duration // 读超时
	WriteWait time.Duration // 写超时
}

type DefaultServer struct {
	Upgrader

	listen string

	ChannelMap
	ServiceRegistration

	Acceptor
	MessageListener
	StateListener

	once sync.Once

	options *ServerOptions
}

func NewServer(listen string, service ServiceRegistration, upgrader Upgrader) *DefaultServer {
	defaultOpts := &ServerOptions{
		LoginWait: DefaultLoginWait,
		ReadWait:  DefaultReadWait,
		WriteWait: DefaultWriteWait,
	}

	return &DefaultServer{
		listen:              listen,
		ServiceRegistration: service,
		Upgrader:            upgrader,
		options:             defaultOpts,
	}
}

func (s *DefaultServer) Start() error {
	log := logger.WithFields(logger.Fields{
		"module": s.Name(),
		"listen": s.listen,
		"id":     s.ServiceID(),
		"func":   "Start",
	})

	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}

	if s.ChannelMap == nil {
		s.ChannelMap = NewChannels(100)
	}

	// 1. 启用连接监听
	lst, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}

	log.Info("started..")

	for {
		// 2. 接收新的连接
		rawconn, err := lst.Accept()
		if err != nil {
			if rawconn != nil {
				_ = rawconn.Close()
			}
			log.Warn(err)
			continue
		}

		go s.connHandler(rawconn)
	}
}

func (s *DefaultServer) connHandler(rawconn net.Conn) {
	rd := pbufio.GetReader(rawconn, ws.DefaultServerReadBufferSize) // todo 应该要配置吧
	wr := pbufio.GetWriter(rawconn, ws.DefaultServerWriteBufferSize)
	defer func() {
		pbufio.PutReader(rd)
		pbufio.PutWriter(wr)
	}()

	conn, err := s.Upgrade(rawconn, rd, wr)
	if err != nil {
		logger.Errorf("Upgrade error: %v", err)
		_ = rawconn.Close()
		return
	}

	// 3. 交给上层处理认证等逻辑
	id, meta, err := s.Accept(conn, s.options.LoginWait)
	if err != nil {
		// 没有通过认证，在关闭当前连接这前，要先通知客户端
		_ = conn.WriteFrame(OpClose, []byte(err.Error()))
		_ = conn.Flush()
		_ = conn.Close()

		// 结束处理当前连接的 goroutine
		return
	}

	if _, ok := s.Get(id); ok {
		_ = conn.WriteFrame(OpClose, []byte("channelId is repeated"))
		_ = conn.Flush()
		_ = conn.Close()

		// 结束处理当前连接的 goroutine
		return
	}

	if meta == nil {
		meta = Meta{}
	}

	// 4. 创建一个 channel 对象，并添加到连接管理中
	channel := NewChannel(id, meta, conn)
	channel.SetReadWait(s.options.ReadWait)
	channel.SetWriteWait(s.options.WriteWait)

	s.Add(channel)

	gaugeWithLabel := channelTotalGauge.WithLabelValues(s.ServiceID(), s.ServiceName())
	gaugeWithLabel.Inc()
	defer gaugeWithLabel.Dec()

	log.Infof("accept user %s in", channel.ID())

	// 5. 循环读取消息，这是一个通用逻辑
	err = channel.ReadLoop(s.MessageListener)
	if err != nil {
		log.Warnf("readloop - ", err)
	}

	// 6. 如果 ReadLoop 返回一个 error，说明连接已经断开，Server 需要把它从 ChannelMap 中删除，并把连接断开事件通知上层
	s.Remove(channel.ID())

	_ = s.Disconnect(channel)

	_ = channel.Close()
}

func (s *DefaultServer) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": s.Name(),
		"id":     s.ServiceID(),
	})

	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()

		for _, ch := range s.ChannelMap.All() {
			_ = ch.Close()
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
	})

	return nil
}

func (s *DefaultServer) Push(id string, data []byte) error {
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return fmt.Errorf("channel %s no found", id)
	}

	return ch.Push(data)
}

func (s *DefaultServer) SetAcceptor(acceptor Acceptor) {
	s.Acceptor = acceptor
}

func (s *DefaultServer) SetMessageListener(listener MessageListener) {
	s.MessageListener = listener
}

func (s *DefaultServer) SetStateListener(listener StateListener) {
	s.StateListener = listener
}

func (s *DefaultServer) SetReadWait(duration time.Duration) {
	s.options.ReadWait = duration
}

func (s *DefaultServer) SetChannelMap(channelMap ChannelMap) {
	s.ChannelMap = channelMap
}

type defaultAcceptor struct {
}

func (d *defaultAcceptor) Accept(conn Conn, timeout time.Duration) (string, Meta, error) {
	return ksuid.New().String(), Meta{}, nil
}
