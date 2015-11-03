package main

import (
    "os"
    _ "fmt"
    "os/signal"
    "syscall"
    "flag"
    "client"
)

var opts client.ClientOption

func init() {
    flag.StringVar(&opts.SrvAddr, "server-addr", "0.0.0.0:7222", "server tcp address")
}

func main() {
    flag.Parse()
    cli := client.NewClient(&opts)

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

    cli.Main()

    <- sigChan
}
