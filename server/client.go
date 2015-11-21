package server

import (
    "net"
//    "io"
    "bufio"
    "fmt"
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

    subChan chan Pub
    subKeys []string

    exitChan chan bool
}

func NewClient(conn net.Conn) *Client {
    return &Client {
        conn: conn,
        reader: bufio.NewReader(conn),
        writer: bufio.NewWriter(conn),
        subChan: make(chan Pub, 1000),
        selectDb: 0,
    }
}

func (c *Client) String() string {
    return c.conn.RemoteAddr().String()
}
func (c *Client) Exit() {
    c.Lock()
    if c.subChan != nil {
        doneFlag := false
        for {
            select{
            case <- c.subChan:
            default:
                doneFlag = true
            }
            if doneFlag == true {
                break
            }
        }
        close(c.subChan)
        c.subChan = nil
    }

    if c.exitChan != nil {
        close(c.exitChan)
    }
    c.Unlock()

    c.Close()
}

func (c *Client) Close() {
    c.conn.Close()
}

func (c *Client) Write(data string) error {
    _, err := c.writer.Write([]byte(data))
    if err != nil {
        return err
    }
    err = c.writer.Flush()
    if err != nil {
        return err
    }
    return nil
}

func (c *Client) ErrorResponse(f string, args...interface{}) error {
    resp := fmt.Sprintf(f, args...)
    _, err := c.writer.Write([]byte(resp))
    if err != nil {
        return err
    }
    err = c.writer.Flush()
    if err != nil {
        return err
    }
    return nil
}
