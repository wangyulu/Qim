package tcp

import (
	"bufio"
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

func (u *Upgrader) Upgrade(rawconn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (kim.Conn, error) {
	conn := NewConnWithRW(rawconn, rd, wr)

	return conn, nil
}
