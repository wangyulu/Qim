package mock

import (
	"errors"
	"time"

	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/tcp"
	"jinv/kim/websocket"
)

type ServerDemo struct {
}

func (s *ServerDemo) Start(id, protocol, addr string) {
	var srv kim.Server

	if protocol == "ws" {
		srv = websocket.NewServer(addr)
	} else if protocol == "tcp" {
		srv = tcp.NewServer(addr)
	}

	handler := &ServerHandler{}

	srv.SetReadWait(time.Minute) // todo 这里设置的超时时间是作用于什么地方
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err := srv.Start()
	if err != nil {
		panic(err)
	}
}

type ServerHandler struct{}

func (h *ServerHandler) Accept(conn kim.Conn, timeout time.Duration) (string, error) {
	// 1. 读取：客户端发送的鉴权数据包
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}

	// 2. 解析：数据包内容就是userId
	userId := string(frame.GetPayload())

	// 3. 鉴权：这里只是为了示例做一个fake验证，非空
	if userId == "" {
		return "", errors.New("user id is invalid")
	}

	return userId, nil
}

func (h *ServerHandler) Receive(ag kim.Agent, payload []byte) {
	ack := string(payload) + " from server"

	_ = ag.Push([]byte(ack))
}

func (h *ServerHandler) Disconnect(id string) error {
	logger.Warnf("disconnect %s", id)

	return nil
}
