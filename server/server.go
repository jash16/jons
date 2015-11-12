package server

import (
    "common"
    "fmt"
    "time"
    "protocol"
    "os"
    "log"
    "net"
    "sync"
)

type cmdFunc func(cli *Client) error

type Server struct {
    db []*JonDb

    tcpListener net.Listener
    Opts *ServerOptions
    logger common.Logger

    cmdMap map[string]cmdFunc

    wg common.WaitGroupWrapper
    sync.Mutex
    exitChan chan bool
}

func NewServer(opt *ServerOptions) *Server {
    return &Server {
        logger: log.New(os.Stderr, "[jon_server]", log.Ldate|log.Ltime|log.Lmicroseconds),
        Opts: opt,
        exitChan: make(chan bool),
        cmdMap: make(map[string]cmdFunc),
    }
}

func (s *Server) Main() {
    var dbs []*JonDb
    for i := 0; i < int(s.Opts.DbNum); i ++ {
        db := NewJonDb()
        dbs = append(dbs, db)
    }
    s.db = dbs

    s.Register()
    tcpListener, err := net.Listen("tcp", s.Opts.TCPAddr)
    if err != nil {
        s.logf("listen %s error - %s", s.Opts.TCPAddr, err)
        os.Exit(1)
    }
    s.tcpListener = tcpListener
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

func (s *Server) initRdb() {

}

func (s *Server) expireLoop() {
    //var expireKeysPerTime int64
    //var expireTimesPerTime int64
    ticker := time.NewTicker(100 * time.Millisecond)
    for {
        select {
        case <- ticker.C:
        }
    }
}

func (s *Server) Exit() {
    if s.tcpListener != nil {
        s.tcpListener.Close()
    }
    close(s.exitChan)
    s.wg.Wait()
}
