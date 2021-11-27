package tcp

import (
	"net"

	"jinv/kim"
)

type Upgrader struct {
}

func NewServer(listen string, service kim.ServiceRegistration) kim.Server {
	return kim.NewServer(listen, service, new(Upgrader))
}

func (u *Upgrader) Name() string {
	return "tcp.server"
}

func (u *Upgrader) Upgrade(rawconn net.Conn) (kim.Conn, error) {
	conn := NewConn(rawconn)

	return conn, nil
}
