package main

import (
    "os"
    "flag"
    "os/signal"
    "syscall"
    "time"
    "server"
)

var opts server.ServerOptions

func init() {
    flag.StringVar(&opts.TCPAddr, "tcp-address","0.0.0.0:7222", "server's listen address")
    flag.DurationVar(&opts.ClientTimeout, "client-timeout", 360 * time.Second, "client timeout")
    flag.IntVar(&opts.DbNum, "-db-num", 16, "server db num")
}

func main() {
    srv := server.NewServer(&opts)

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

    srv.Main()

    <- sigChan
    //srv.logf("jon_server quit")
}
