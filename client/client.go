package client

import (
    "io"
    "os"
    "fmt"
    "net"
    "time"
    "sync"
    "bufio"
    "bytes"
    "common"
    "strconv"
    _ "io/ioutil"
)

var ClientPromot string = "not connected> "
var FailConnect string

const (
    done int = iota
    needMore
)
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
        fmt.Printf("%s\n", FailConnect)
        c.connected = false
    } else {
        c.conn = conn
        c.reader = bufio.NewReader(conn)
        c.writer = bufio.NewWriter(conn)

        c.connected = true
        ClientPromot = c.opts.SrvAddr + "> "
    }
    c.wg.Wrap(func(){
        c.ioLoop()
    })

    c.wg.Wrap(func(){
        c.connectLoop()
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
//        FailConnect = fmt.Sprintf("Could not connect to Redis at %s: %s", c.opts.SrvAddr, err)
//        fmt.Printf("%s\n", FailConnect)
        c.connected = false
        ClientPromot = "not connected> "
        FailConnect = fmt.Sprintf("Could not connect to Redis at %s: %s", c.opts.SrvAddr, err)
        return err
    }
    c.conn = conn
    c.reader = bufio.NewReader(conn)
    c.writer = bufio.NewWriter(conn)
    c.connected = true
    ClientPromot = c.opts.SrvAddr + "> "
    return nil
}

func (c *Client) connectLoop() {
    ticker := time.NewTicker(5 * time.Second)
    for {
        select {
        case <- ticker.C:
            c.Connect()
        }
    }
}

func (c *Client) ioLoop() {
    //var respFlag int
    //var respData []byte
    stdReader := bufio.NewReader(os.Stdin)
//    ioReader := c.reader
//    ioWriter := c.writer
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
        ioReader := c.reader
        ioWriter := c.writer
        data := c.Construct(cmd)
        /*write request*/
        _, err = ioWriter.WriteString(data)
        if err != nil {
            FailConnect = fmt.Sprintf("Could not connect to Redis at %s: %s", c.opts.SrvAddr, err)
            c.connected = false
            continue
        //    continue
        }
        err = ioWriter.Flush()
        if err != nil {
            FailConnect = fmt.Sprintf("Could not connect to Redis at %s: %s", c.opts.SrvAddr, err)
            c.connected = false
            continue
        }
        /*read response*/
        resp, err := ioReader.ReadSlice('\n')
        if err != nil {
            FailConnect = fmt.Sprintf("Could not connect to Redis at %s: %s", c.opts.SrvAddr, err)
            c.connected = false
            continue
        }
        err = c.processResponse(resp)
        if err != nil {
            FailConnect = fmt.Sprintf("Could not connect to Redis at %s: %s", c.opts.SrvAddr, err)
            c.connected = false
            continue
        }
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

func (c *Client) processResponse(resp []byte) error {
    var firstChar byte
    var respData []byte
    var err error
    //var i int
    dataLen := len(resp)

    if dataLen < 1 {
        return err
    }

//    fmt.Printf("%s", resp)
    firstChar = resp[0]
    switch firstChar {
    case '+':
        respData = resp[1: dataLen]
        fmt.Printf("%s", respData)
        break
    case '-':
        respData = resp[1: dataLen]
        fmt.Printf("(error) %s", respData)
        break
    case '$':
        respData = resp[1: dataLen - 2] //remove \r\n
        dataLen, err := strconv.Atoi(string(respData))
        if err != nil {
            return err
        }
        //fmt.Printf("%d\n", dataLen)
        if dataLen <= 0 {
            fmt.Printf("(nil)\n")
            break
        }
        leftData := make([]byte, dataLen + 2)
        _, err = io.ReadFull(c.reader, leftData)
        if err != nil {
            return err
        }
        fmt.Printf("%s", leftData)
    case '*':
        //fmt.Printf("left Line: %s", resp)
        respData = resp[1: dataLen - 2]
        leftLines, err := strconv.Atoi(string(respData))
        if err != nil {
            return err
        }
        //fmt.Printf("left lines: %d\n", leftLines)
        for i := 0; i < leftLines; i ++ {
            _, err := c.reader.ReadSlice('\n')
            if err != nil {
                return err
            }
          //  fmt.Printf("left data: %d", dataLen)
            data, err := c.reader.ReadSlice('\n')
            if err != nil {
                return err
            }
            fmt.Printf("%d) %s", i, data)
        }
    }
    return err
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
