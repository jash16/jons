package client

import (
    "os"
    "fmt"
    "net"
    "time"
    "sync"
    "bufio"
    "bytes"
    "common"
)

var ClientPromot string = "not connected> "
var FailConnect string

type Client struct {
    opts *ClientOption
    conn net.Conn
    reader *bufio.Reader
    writer *bufio.Writer
    connected bool

    sync.Mutex

    wg common.WaitGroupWrapper
    exitChan chan bool
}

func NewClient(opt *ClientOption) *Client {
    return &Client {
        opts: opt,
        exitChan: make(chan bool),
        connected: false,
    }
}

func (c *Client) Main() {
    dialer := &net.Dialer {
        Timeout: 5 * time.Second,
    }
    conn, err := dialer.Dial("tcp", c.opts.SrvAddr)
    if err != nil {
        FailConnect = fmt.Sprintf("Could not connect to Redis at %s: %s", c.opts.SrvAddr, err)
        fmt.Printf("%s", FailConnect)
        c.connected = false
    }

    c.conn = conn
    c.reader = bufio.NewReader(conn)
    c.writer = bufio.NewWriter(conn)

    c.connected = true

    ClientPromot = c.opts.SrvAddr + "> "
    c.wg.Wrap(func(){
        c.ioLoop()
    })
}

func (c *Client) Connect() error {
    if c.connected == true {
        return nil
    }

    dialer := &net.Dialer {
        Timeout: 5 * time.Second,
    }

    conn, err := dialer.Dial("tcp", c.opts.SrvAddr)
    if err != nil {
        FailConnect = fmt.Sprintf("Could not connect to Redis at %s: %s", c.opts.SrvAddr, err)
        fmt.Printf("%s", FailConnect)
        c.connected = false
        return err
    }
    c.conn = conn
    c.reader = bufio.NewReader(conn)
    c.writer = bufio.NewWriter(conn)
    c.connected = true
    return nil
}

func (c *Client) ioLoop() {
    stdReader := bufio.NewReader(os.Stdin)
    ioReader := c.reader
    ioWriter := c.writer
    for {
        fmt.Printf("%s", ClientPromot)
        cmd, err := stdReader.ReadSlice('\n')
        if err != nil {
            break
        }
        if c.connected == false {
            fmt.Printf("%s\n", FailConnect)
            continue
        }
        data := c.Construct(cmd)
        /*write request*/
        _, err = ioWriter.WriteString(data)
        if err != nil {
            continue
        }
        err = ioWriter.Flush()
        if err != nil {
            continue
        }
        /*read response*/
        resp, err := ioReader.ReadSlice('\n')
        if err != nil {
            break
        }
        c.processRespone(resp)
        //fmt.Printf("%s", resp)
    }
}

func (c *Client)Construct(raw []byte) string {
    var data string
    raw = bytes.Trim(raw, " \n")
    rawParams := bytes.Split(raw, []byte(" "))
    paramNum := len(rawParams)
    data = fmt.Sprintf("*%d\r\n", paramNum)
    for _, param := range rawParams {
        paramLen := len(param)
        data += fmt.Sprintf("$%d\r\n", paramLen)
        data += fmt.Sprintf("%s\r\n", param)
    }
    return data
}

func (c *Client) processResponse(resp []byte) {

}

func (c *Client) Exit() {
    c.Lock()
    if c.conn != nil {
        c.conn.Close()
    }
    c.Unlock()

    close(c.exitChan)

    c.wg.Wait()
}
