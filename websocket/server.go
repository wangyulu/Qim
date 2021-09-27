package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/segmentio/ksuid"
	"jinv/kim"
	"jinv/kim/logger"
)

type ServerOptions struct {
	loginWait time.Duration
	readWait  time.Duration
	writeWait time.Duration
}

type Server struct {
	listen string

	kim.ChannelMap

	kim.Acceptor
	kim.MessageListener
	kim.StateListener

	once sync.Once

	options ServerOptions
}

func NewServer(listen string) kim.Server {
	return &Server{
		listen: listen,
		options: ServerOptions{
			loginWait: kim.DefaultLoginWait,
			readWait:  kim.DefaultReadWait,
			writeWait: kim.DefaultWriteWait * 10,
		},
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	log := logger.WithFields(logger.Fields{
		"module": "ws.server",
		"listen": s.listen,
	})

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}

	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}

	// 连接管理器
	if s.ChannelMap == nil {
		s.ChannelMap = kim.NewChannels(100) // todo num
	}

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// 1.建立连接
		rawconn, _, _, err := ws.UpgradeHTTP(request, writer)
		if err != nil {
			resp(writer, http.StatusBadRequest, err.Error())
			return
		}

		// 2.包装 conn
		conn := NewConn(rawconn)

		// 3. 回调给上层业务完成权限认证之类的逻辑处理
		id, err := s.Accept(conn, s.options.loginWait)
		if err != nil {
			_ = conn.WriteFrame(kim.OpClose, []byte(err.Error()))
			conn.Close()
			return
		}

		if _, ok := s.Get(id); ok {
			log.Warnf("channel %s existed", id)
			_ = conn.WriteFrame(kim.OpClose, []byte("channcelId is repeated"))
			conn.Close()
			return
		}

		// 4. 创建 Channel，自动添加到kim.ChannelMap连接管理器中
		channel := kim.NewChannel(id, conn)
		channel.SetWriteWait(s.options.writeWait)
		channel.SetReadWait(s.options.readWait)

		s.Add(channel)

		log.Infof("accept user %s in", channel.ID())

		go func(ch kim.Channel) {
			// 5. 开启一个goroutine中循环读取消息。这里是调用了Channel中的Readloop方法
			err := ch.ReadLoop(s.MessageListener)
			if err != nil {
				log.Warn("readloop - ", err)
			}

			// 6.
			s.Remove(ch.ID())

			err = s.Disconnect(ch.ID())
			if err != nil {
				log.Warn(err)
			}
			ch.Close()
		}(channel)
	})

	log.Infoln("started..")

	return http.ListenAndServe(s.listen, mux)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.once.Do(func() {
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

func resp(w http.ResponseWriter, code int, body string) {
	w.WriteHeader(code)
	if body != "" {
		_, _ = w.Write([]byte(body))
	}
	logger.Warnf("response with code:%d %s", code, body)
}

type defaultAcceptor struct {
}

func (a *defaultAcceptor) Accept(conn kim.Conn, duration time.Duration) (string, error) {
	return ksuid.New().String(), nil
}
