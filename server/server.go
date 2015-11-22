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

    //for sub pub
    subMap map[string][]*Client
    subLock sync.RWMutex
    subExitChan chan subExit

    rdbFlag bool
    rdbHandler *os.File
    //persist
    p Persist
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
        subMap: make(map[string][]*Client),
        subExitChan: make(chan subExit),
      //  p: &rdb{},
        rdbFlag: false,
    }
}

func (s *Server) Main() {
    var dbs []*JonDb
    for i := 0; i < int(s.Opts.DbNum); i ++ {
        db := NewJonDb()
        dbs = append(dbs, db)
    }
    s.db = dbs

    //register funcs
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

    p := &rdb {
        ctx: ctx,
    }
    s.p = p

    s.initRdb()

    s.wg.Wrap(func() {
        protocol.TCPServer(s.tcpListener, tcpSrv, s.logger)
    })
    s.wg.Wrap(func() {
        s.pubsubLoop()
    })
}

func (s *Server) logf(data string, args...interface{}) {
    s.logger.Output(2, fmt.Sprintf(data, args...))
}

func (s *Server) initRdb() error {
    err := s.p.Load("dump.rdb")
    if err != nil {
        s.logf("load dump.rdb failed - %s", err)
    }
    return err
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

func (s *Server) pubsubLoop() {
    //var clis []*Client
    var nclis []*Client
    for {
        select {
        case subExitCli := <- s.subExitChan:
            key := subExitCli.key
            cli := subExitCli.cli
            s.logf("subscribe client: %s quit", cli)
            s.subLock.Lock()
            if clis, ok := s.subMap[key]; ! ok {
                continue
            } else {
                for _, c := range clis {
                    if c != cli {
                        nclis = append(nclis, c)
                    }
                }
                s.subMap[key] = nclis
            }
            s.subLock.Unlock()
        }
    }
}

func (s *Server) AddSubClient(subkey string, cli *Client) {
    s.subLock.Lock()
    defer s.subLock.Unlock()
    //var clis []*Client
    if clis, ok := s.subMap[subkey]; ! ok {
        clis = append(clis, cli)
        s.subMap[subkey] = clis
    } else {
        for _, c := range clis {
            if c == cli {
                return
            }
        }
        clis = append(clis, cli)
        s.subMap[subkey] = clis
    }
    return
}

func (s *Server) Exit() {
    s.Lock()
    if s.tcpListener != nil {
        s.tcpListener.Close()
    }
    s.Unlock()
    close(s.exitChan)
    s.wg.Wait()
}
