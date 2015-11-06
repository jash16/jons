package server

import (
    "common"
    "protocol"
    "os"
    "log"
    "net"
    "sync"
)

type Server struct {
    db []*JonDb

    tcpListener *net.Listener
    Opts *ServerOption
    logger Logger

    wg common.WaitGroupWrapper
    sync.Mutex
    exitChan chan bool
}

func NewServer(opt *ServerOption) *Server {
    return &Server {
        logger: log.New(os.Stderr, "[jon_server]", log.Ldate|log.Ltime|log.Lmicroseconds),
        Opts: opt,
        exitChan: make(chan bool)
    }
}

func (s *Server) Main() {
    var dbs []*JobDb
    for i := 0; i < s.Opts.DbNum; i ++ {
        db := NewJonDb()
        dbs = append(dbs, db)
    }
    s.db = dbs

    tcpListener, err := net.Listen("tcp", s.Opts.TCPAddr)
    if err != nil {
        s.logf("listen %s error - %s", s.Opts.TCPAddr, err)
        os.Exit(1)
    }
    s.tcpListener = tcpListner
    ctx := &context{s: s}
    tcpSrv := &tcpServer {
        ctx: ctx,
    }

    s.wg.Wrap(func() {
        protocol.TCPServer(s.tcpListener, tcpSrv, s.logger)
    })
}

func (s *Server) logf(data string, args...interface{}) {
    s.logger.Output(2, fmt.Sprintf(data, args...))
}

func (s *Server) Exit() {
    if s.tcpListener != nil {
        s.tcpListener.Close()
    }
    close(s.exitChan)
    s.Wait()
}
