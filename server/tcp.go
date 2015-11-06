package server

import (
    "net"
)

type tcpServer struct {
    ctx *context
}

func (t *tcpServer)Handle(conn net.Conn) {
    t.ctx.s.logf("receive connect: %s", conn.RemoteAddr())

}

func (t *tcpServer)IOLoop() {

}
