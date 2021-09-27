package mock

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/tcp"
	"jinv/kim/websocket"
)

type ClientDemo struct {
}

func (c *ClientDemo) Start(userId, protocol, addr string) {
	var cli kim.Client

	// 1. 初始化客户端
	if protocol == "ws" {
		cli = websocket.NewClient(userId, "client", websocket.ClientOptions{})
		// 2. 设置拨号器
		cli.SetDialer(&WebsocketDialer{})
	} else if protocol == "tcp" {
		cli = tcp.NewClient(userId, "client", tcp.ClientOptions{})
		cli.SetDialer(&TCPDialer{})
	}

	// 3. 建立连接
	err := cli.Connect(addr)
	if err != nil {
		logger.Error(err)
	}

	count := 10

	// 5. 开启 goroutine，每隔一秒向服务器端发送1条消息，发送10次后退出
	go func() {
		for i := 0; i < count; i++ {
			err := cli.Send([]byte(fmt.Sprintf("hello_%d", i)))
			if err != nil {
				logger.Error(err)
				return
			}

			time.Sleep(time.Second)
		}
	}()

	// 6. 读取消息，并在读取完10条消息后退出
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Info("client read ", err)
			break
		}

		if frame.GetOpCode() != kim.OpBinary {
			continue
		}

		recv++

		logger.Warnf("%s receive message [%s]", cli.ID(), frame.GetPayload())

		if recv == count {
			break
		}
	}

	// 7. 退出
	cli.Close()
}

type WebsocketDialer struct{}

func (d *WebsocketDialer) DialAndHandshake(ctx kim.DialerContext) (net.Conn, error) {
	var cancelFunc context.CancelFunc

	durationCtx := context.Background()

	// todo 如何调试这里的超时时间是否有作用呢
	if ctx.Timeout > 0 {
		durationCtx, cancelFunc = context.WithDeadline(context.Background(), time.Now().Add(ctx.Timeout))
		defer cancelFunc()
	}

	// 1. 拨号
	conn, _, _, err := ws.Dial(durationCtx, ctx.Address)

	if err != nil {
		return nil, err
	}

	// 2. 发送用户认证，示例是 userId
	err = wsutil.WriteClientBinary(conn, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

type TCPDialer struct{}

func (d *TCPDialer) DialAndHandshake(ctx kim.DialerContext) (net.Conn, error) {
	// 1. 拨号
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}

	// 2. 发送用户认证，示例是 userId
	err = tcp.WriteFrame(conn, kim.OpBinary, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}

	// 3. 返回连接
	return conn, nil
}
