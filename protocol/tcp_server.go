package protocol

import (
    "net"
    "fmt"
    "strings"
    "runtime"
    "common"
)

type TCPHandler interface {
    Handle(conn net.Conn)
}

func TCPServer(listener net.Listener, handler TCPHandler, l common.Logger) {
    l.Output(2, fmt.Sprintf("TCP: listening on %s", listener.Addr()))

    for {
        clientConn, err := listener.Accept()
        if err != nil {
            if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
                l.Output(2, fmt.Sprintf("Notice: temporary accept failure - %s", err.Error()))
                runtime.Gosched()
                continue
            }
            //not closed 
            if !strings.Contains(err.Error(), "use of closed network connection") {
                l.Output(2, fmt.Sprintf("Error: listener.Accpet() - %s", listener.Addr()))
            }
            break
        }

        go handler.Handle(clientConn)
    }
}
