package server

import (
    "common"
    "log"
    "net"
)

type Server struct {
    db []*JonDb

    tcpListener *net.Listener
    Opts *ServerOption
    logger Logger

    exitChan chan bool
}

func NewServer(opt *ServerOption) *Server {
    return &Server {
        logger: log.New(os.Stderr, "[jon_server]", log.Ldate|log.Ltime|log.Lmicroseconds),
        Opts: opt,
    }
}

func (s *Server) Main() {

}

func (s *Server) logf(data string, args...interface{}) {
    s.logger.Output(2, fmt.Sprintf(data, args...))
}

func Exit() {
}
