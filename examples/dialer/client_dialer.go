package dialer

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/wire"
	"jinv/kim/wire/pkt"
	"jinv/kim/wire/token"
)

type ClientDialer struct {
	AppSecret string
}

func (d *ClientDialer) DialAndHandshake(ctx kim.DialerContext) (net.Conn, error) {
	logger.Info("DialAndHandshake call")

	// 1. 拨号
	conn, _, _, err := ws.Dial(context.Background(), ctx.Address)
	if err != nil {
		return nil, err
	}

	if d.AppSecret == "" {
		d.AppSecret = token.DefaultSecret
	}

	// 2. 直接使用JWT生成一个Token
	tk, err := token.Generate(d.AppSecret, &token.Token{
		Account: ctx.Id,
		App:     "kim",
		Exp:     time.Now().AddDate(0, 0, 1).Unix(),
	})
	if err != nil {
		return nil, err
	}

	// 3. 发送一条CommandLoginSignIn消息
	loginReq := pkt.New(wire.CommandLoginSignIn).WriteBody(&pkt.LoginReq{
		Token: tk,
	})

	if err := wsutil.WriteClientBinary(conn, pkt.Marshal(loginReq)); err != nil {
		return nil, err
	}

	logger.Info("waiting for login response")

	_ = conn.SetReadDeadline(time.Now().Add(ctx.Timeout))

	frame, err := ws.ReadFrame(conn)
	if err != nil {
		return nil, err
	}

	ack, err := pkt.MustReadLogicPkt(bytes.NewBuffer(frame.Payload))
	if err != nil {
		return nil, fmt.Errorf("ack err ", err)
	}

	// 4. 判断是否登录成功
	if ack.Status != pkt.Status_Success {
		return nil, fmt.Errorf("login failed: %v", &ack.Header)
	}

	var resp = new(pkt.LoginResp)
	_ = ack.ReadBody(resp)
	logger.Info("login ", resp.GetChannelId())

	return conn, nil
}
