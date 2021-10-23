package serv

import (
	"net"

	"google.golang.org/protobuf/proto"
	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/tcp"
	"jinv/kim/wire/pkt"
)

type TcpDialer struct {
	ServiceId string
}

func NewDialer(serviceId string) kim.Dialer {
	return &TcpDialer{
		ServiceId: serviceId,
	}
}

func (d *TcpDialer) DialAndHandshake(ctx kim.DialerContext) (net.Conn, error) {
	// 1. 拨号建立连接
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}

	req := &pkt.InnerHandshakeReq{
		ServiceId: d.ServiceId,
	}

	logger.Infof("send req %v", req)

	// 2. 把自己的 serviceId 发送给对方
	bts, err := proto.Marshal(req)

	if err = tcp.WriteFrame(conn, kim.OpBinary, bts); err != nil {
		return nil, err
	}

	return conn, nil
}
