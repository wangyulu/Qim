package websocket

import (
	"net"

	"github.com/gobwas/ws"
	"jinv/kim"
)

type Upgrader struct {
}

func NewServer(listen string, service kim.ServiceRegistration) kim.Server {
	return kim.NewServer(listen, service, new(Upgrader))
}

func (u *Upgrader) Name() string {
	return "websocket.server"
}

func (u *Upgrader) Upgrade(rawconn net.Conn) (kim.Conn, error) {
	_, err := ws.Upgrade(rawconn)
	if err != nil {
		return nil, err
	}

	conn := NewConn(rawconn)

	return conn, nil
}
