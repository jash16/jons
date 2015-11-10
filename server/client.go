package server

import (
    "net"
    "io"
    "bufio"
    "sync"
)

type Client struct {
    argc int32
    argv [][]byte
    conn net.Conn
    reader *bufio.Reader
    writer *bufio.Writer

    sync.Mutex

    selectDb int32
    db *JonDb
    respBuf []byte

    exitChan chan bool
}

func NewClient(conn net.Conn) *Client {
    return &Client {
        conn: conn,
        reader: bufio.NewReader(conn),
        writer: bufio.NewWriter(conn),
        selectDb: 0,
    }
}

func (c *Client) Exit() {
    c.Lock()
    if c.exitChan != nil {
        close(c.exitChan)
    }
    c.Unlock()
    c.Close()
}

func (c *Client) Close() {
    c.conn.Close()
}
