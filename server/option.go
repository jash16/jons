package server

import (
    "time"
)

type ServerOptions struct {
    TCPAddr string
    ClientTimeout time.Duration
    DbNum int32
}
