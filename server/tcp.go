package server

import (
    "net"
)

type tcpServer struct {
    ctx *context
}

func (t *tcpServer)Handle(conn net.Conn) {
    t.ctx.s.logf("receive connect: %s", conn.RemoteAddr())
    proto := &JonProtocol{
        ctx: t.ctx,
    }

    err := proto.IOLoop(conn)
    if err != nil {
        t.ctx.s.logf("client %s err - %s", conn.RemoteAddr(), err)
    }
}
