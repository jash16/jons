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
    "sync/atomic"
)

type Role int32

const (
    Master Role = 0
    Slave Role = 1
)

type Server struct {
    db []*JonDb

    tcpListener net.Listener
    Opts *ServerOptions
    logger common.Logger

    cmdMap map[string]cmdFunc

    //for replication
    slaves []*Client
    role   Role
    masterAddr string

    //for sub pub
    subMap map[string][]*Client
    subLock sync.RWMutex
    subExitChan chan subExit

    rdbFlag bool
    //rdbHandler *os.File
    //persist
    p Persist

    //for aof
    //aofSelectDb int
    aof *aof
    dc chan dirtyCmd
    cmds []dirtyCmd
    aofFlag int32

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
        dc: make(chan dirtyCmd, 5000),
        rdbFlag: false,
        aofFlag: 0,
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

    s.aof = &aof{aofSelectDb: -10,}
    s.initRdb()

    s.wg.Wrap(func() {
        protocol.TCPServer(s.tcpListener, tcpSrv, s.logger)
    })
    s.wg.Wrap(func() {
        s.pubsubLoop()
    })
    s.wg.Wrap(func() {
        s.srvLoop()
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

func (s *Server) srvLoop() {
    expireTicker := time.NewTicker(1 * time.Second)
    aofTicker := time.NewTicker(1 * time.Second)
    rdbTicker := time.NewTicker(5 * time.Second)
    for {
        select {
        case cmd := <- s.dc:
            s.logf("receive command")
            s.cmds = append(s.cmds, cmd)
            if s.Opts.SyncEvery && atomic.CompareAndSwapInt32(&s.aofFlag, 0, 1) {
                cmds := make([]dirtyCmd, len(s.cmds))
                copy(cmds, s.cmds)
                err := s.aof.appendCmdSToFile(cmds)
                if err != nil {
                    s.logf("append only file: %s", err)
                    s.cmds = cmds
                }
                atomic.CompareAndSwapInt32(&s.aofFlag, 1, 0)
            }
        case <- aofTicker.C:
            if len(s.cmds) > 0 && atomic.CompareAndSwapInt32(&s.aofFlag, 0, 1) {
                s.logf("in here")
                cmds := make([]dirtyCmd, len(s.cmds))
                copy(cmds, s.cmds)
                s.cmds = s.cmds[:0]
                err := s.aof.appendCmdSToFile(cmds)
                if err != nil {
                    s.logf("append only file: %s", err)
                    for _, val := range s.cmds {
                        cmds = append(cmds, val)
                    }
                    s.cmds = cmds
                }
                atomic.CompareAndSwapInt32(&s.aofFlag, 1, 0)
                //s.aofFlag = 0
            }
        case <- rdbTicker.C:
        case <- expireTicker.C:

        case <- s.exitChan:
            goto exit
        }
    }
exit:
    s.logf("srvLoop done!")
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
