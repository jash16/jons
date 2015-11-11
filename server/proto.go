package server

import (
    "io"
    "fmt"
    "net"
    "strconv"
)

var dump []byte

func init() {
    defaultSize := 512 * 1024 * 1024 + 512
    dump = make([]byte, defaultSize)
}

type Protocol interface {
    IOLoop(conn net.Conn) error
}

type JonProtocol struct {
    ctx *context
}

func (p *JonProtocol)IOLoop(conn net.Conn) error {
    var srvExitChan chan bool
    var cliExitChan chan bool
    var err error
    var cmdNum, nextNum int
    var line []byte
    var cmd []byte
    srvExitChan = p.ctx.s.exitChan
    client := NewClient(conn)
    //client.db = p.ctx.s.db[client.selectDb]
    r := client.reader
   // w := client.writer
    cliExitChan = client.exitChan
    for {
        select {
        case <- cliExitChan:
            goto end
        case <- srvExitChan:
            goto end
        default:
        }
        client.argc = 0
        client.argv = client.argv[0:0]
        line, err = r.ReadSlice('\n')
        if err != nil {
            goto end
        }
        if len(line) <= 3 || line[0] != '*' {
            io.ReadFull(r, dump)
            client.ErrorResponse("ERR unknown command '%s'", line)
            continue
        }
        cmdNum, err = strconv.Atoi(string(line[1:len(line)-2]))
        if err != nil {
            io.ReadFull(r, dump)
            client.ErrorResponse("ERR unknown command '%s'", line)
            continue
        }
        for i := 0; i < cmdNum; i ++ {
            line, err = r.ReadSlice('\n')
            if err != nil {
                io.ReadFull(r, dump)
                client.ErrorResponse("ERR unknown command '%s'", line)
                break
            }
            if len(line) < 4 || line[0] != '$' {
                io.ReadFull(r, dump)
                client.ErrorResponse("ERR unknown command '%s'", line)
                break
            }
            nextNum, err = strconv.Atoi(string(line[1:len(line)-2]))
            if err != nil {
                goto ERR
            }
            cmd = make([]byte, nextNum+2)
            io.ReadFull(r, cmd)
            client.argv = append(client.argv, cmd[0:len(cmd)-2])
            client.argc ++;
        }
        err = p.processCommand(client)
        if err != nil {
            goto ERR
        } else {
            continue
        }
ERR:
        io.ReadFull(r, dump)
        client.ErrorResponse("ERR unknown command '%s'", line)
    }
end:
    client.Exit()
    return err
}

func (p *JonProtocol) processCommand(cli *Client) error {
    var err error
    switch string(cli.argv[0]) {
    case "set":
        err = p.Set(cli)
    case "select":
        err = p.Select(cli)
    case "get":
        err = p.Get(cli)
    default:
        cli.ErrorResponse("ERR unknown command '%s'", cli.argv[0])
    }
    return err
}

func (p *JonProtocol) Set(cli *Client) error {
    if (cli.argc != 3) {
        cli.ErrorResponse("ERR wrong number of arguments for 'set' command")
        return nil
    }
    key_str := string(cli.argv[1])
    val_str := string(cli.argv[2])
    p.ctx.s.logf("receive command: %s %s %s", cli.argv[0], cli.argv[1], cli.argv[2])
    /*
    K := Key {
        //Type: JON_STRING,
        Ref: 1,
        Value: key_str,
    }
    */
    V := NewElement(JON_STRING, val_str)
    p.ctx.s.Lock()
    defer p.ctx.s.Unlock()
    db := p.ctx.s.db[cli.selectDb]
    if val, ok := db.Dict.DataMap[key_str]; ok {
        if val.Type != JON_STRING {
            cli.Write("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")
            return nil
        }
    }
    db.Dict.DataMap[key_str] = V
    cli.Write("+OK\r\n")
    return nil
}

func (p *JonProtocol) Select(cli *Client) error {
    cli.Write("+OK\r\n")
    return nil
}

func (p *JonProtocol) Get(cli *Client) error {
     var val *Element
     var ok bool
     if cli.argc != 2 {
         cli.ErrorResponse("ERR wrong number of arguments for 'set' command")
         return nil
     }
     p.ctx.s.Lock()
     defer p.ctx.s.Unlock()
     key := string(cli.argv[1])
     db := p.ctx.s.db[cli.selectDb]
     if val, ok = db.Dict.DataMap[key]; !ok {
         cli.Write("$-1\r\n")
         return nil
     }
     if val.Type != JON_STRING {
         cli.Write("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")
         return nil
     }
     valStr, _ := val.Value.(string)
     valLen := len(valStr)
     respStr := fmt.Sprintf("$%d\r\n%s\r\n", valLen, valStr)
     cli.Write(respStr)
     return nil
}
