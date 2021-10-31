package tcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/segmentio/ksuid"
	"jinv/kim"
	"jinv/kim/logger"
)

type ServerOptions struct {
	loginWait time.Duration // 登录超时
	readWait  time.Duration // 读超时
	writeWait time.Duration // 写超时
}

type Server struct {
	listen string

	kim.ChannelMap
	kim.ServiceRegistration

	kim.Acceptor
	kim.MessageListener
	kim.StateListener

	once sync.Once

	options ServerOptions

	quit *kim.Event
}

func NewServer(listen string, service kim.ServiceRegistration) kim.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		ChannelMap:          kim.NewChannels(100),
		quit:                kim.NewEvent(),
		options: ServerOptions{
			loginWait: kim.DefaultLoginWait,
			readWait:  kim.DefaultReadWait,
			writeWait: kim.DefaultWriteWait,
		},
	}
}

func (s *Server) Start() error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}

	// 1. 启用连接监听
	lst, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}

	log.Info("started")

	for {
		// 2. 接收新的连接
		rawconn, err := lst.Accept()
		if err != nil {
			rawconn.Close() // todo 为什么这里在有错误的情况下还要关闭，既然关闭为何没有通知客户端呢
			log.Warn("accept - ", err)
			continue
		}

		go func(rawconn net.Conn) {
			conn := NewConn(rawconn)

			// 3. 交给上层处理认证等逻辑
			id, err := s.Accept(conn, s.options.loginWait)
			if err != nil {
				// 没有通过认证，在关闭当前连接之前，要先通知客户端
				_ = conn.WriteFrame(kim.OpClose, []byte(err.Error()))
				conn.Close()
				// 结束处理当前链接的 goroutine
				return
			}

			if _, ok := s.Get(id); ok {
				log.Warnf("channel %s existed", id)
				_ = conn.WriteFrame(kim.OpClose, []byte("channelId is repeated"))
				conn.Close()
				// 结束处理当前链接的 goroutine
				return
			}

			// 4. 创建一个 channel 对象，并添加到连接管理中
			channel := kim.NewChannel(id, conn)
			channel.SetReadWait(s.options.readWait)
			channel.SetWriteWait(s.options.writeWait)

			s.Add(channel)

			log.Infof("accept user %s in", channel.ID())

			// 5. 循环读取消息，这是一个通用逻辑
			err = channel.ReadLoop(s.MessageListener)
			if err != nil {
				log.Warnf("readloop - ", err)
			}

			// 6. 如果 ReadLoop 返回一个 error，说明连接已经断开，Server需要把它从 channelMap 中删除，并把连接断开事件通知上层
			s.Remove(channel.ID())

			_ = s.Disconnect(channel.ID())

			channel.Close()
		}(rawconn)

		// todo 这里的作用是什么呢
		select {
		case <-s.quit.Done():
			return fmt.Errorf("listen exited")
		default:

		}
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"id":     s.ServiceID(),
	})

	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()

		for _, ch := range s.ChannelMap.All() {
			ch.Close()
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

func (s *Server) Push(id string, data []byte) error {
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return fmt.Errorf("channel %s no found", id)
	}

	return ch.Push(data)
}

func (s *Server) SetAcceptor(acceptor kim.Acceptor) {
	s.Acceptor = acceptor
}

func (s *Server) SetMessageListener(listener kim.MessageListener) {
	s.MessageListener = listener
}

func (s *Server) SetStateListener(listener kim.StateListener) {
	s.StateListener = listener
}

func (s *Server) SetReadWait(duration time.Duration) {
	s.options.readWait = duration
}

func (s *Server) SetChannelMap(channelMap kim.ChannelMap) {
	s.ChannelMap = channelMap
}

type defaultAcceptor struct {
}

func (d *defaultAcceptor) Accept(conn kim.Conn, timeout time.Duration) (string, error) {
	return ksuid.New().String(), nil
}
