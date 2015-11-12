package server

import (
    "io"
    //"fmt"
    "net"
    "strconv"
)

var dump []byte

func init() {
    defaultSize := 512 * 1024 * 1024 + 512
    dump = make([]byte, defaultSize)
}

var (
    ok string = "+OK\r\n"
    wrongType string = "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
    wrongArgs string = "-ERR wrong number of arguments for '%s' command\r\n"
    wrongCommand string = "-ERR unknown command '%s'\r\n"
    wrongDbIdx string = "-ERR invalid DB index\r\n"
    wrongKey string = "$-1\r\n"
    wrongIdx string = "-ERR index out of range\r\n"
)

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
    r := client.reader
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
            client.ErrorResponse(wrongCommand, line)
            continue
        }
        cmdNum, err = strconv.Atoi(string(line[1:len(line)-2]))
        if err != nil {
            io.ReadFull(r, dump)
            client.ErrorResponse(wrongCommand, line)
            continue
        }
        for i := 0; i < cmdNum; i ++ {
            line, err = r.ReadSlice('\n')
            if err != nil {
                io.ReadFull(r, dump)
                client.ErrorResponse(wrongCommand, line)
                break
            }
            if len(line) < 4 || line[0] != '$' {
                io.ReadFull(r, dump)
                client.ErrorResponse(wrongCommand, line)
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
        client.ErrorResponse(wrongCommand, line)
    }
end:
    client.Exit()
    return err
}

func (p *JonProtocol) processCommand(cli *Client) error {

    var err error
    cmd := string(cli.argv[0])
    if cmdFunc, ok := p.ctx.s.cmdMap[cmd] ; ok {
        err = cmdFunc(cli)
    } else {
        cli.ErrorResponse(wrongCommand, cli.argv[0])
    }
    return err
}

/*
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
        cli.ErrorResponse(wrongCommand, cli.argv[0])
    }
    return err
}

func (p *JonProtocol) Set(cli *Client) error {
    if (cli.argc != 3) {
        cli.ErrorResponse(wrongArgs, "set")
        return nil
    }
    key_str := string(cli.argv[1])
    val_str := string(cli.argv[2])
    p.ctx.s.logf("receive command: %s %s %s", cli.argv[0], cli.argv[1], cli.argv[2])
    K := key_str
    V := NewElement(JON_STRING, val_str)
    db := p.ctx.s.db[cli.selectDb]
    typ := db.LookupKeyType(K)
    if typ != JON_KEY_NOTEXIST && typ != JON_STRING {
        cli.Write(wrongType)
        return nil
    }
    db.SetKey(K, V)
    cli.Write(ok)
    return nil
}

func (p *JonProtocol) Select(cli *Client) error {
    if cli.argc != 2 {
        cli.ErrorResponse(wrongArgs, "select")
        return nil
    }
    dbIdx, _ := strconv.Atoi(string(cli.argv[1]))
    if dbIdx >= 16 || dbIdx < 0 {
        cli.ErrorResponse(wrongDbIdx)
        return nil
    }
    cli.selectDb = int32(dbIdx)
    cli.Write(ok)
    return nil
}

func (p *JonProtocol) Get(cli *Client) error {
     var val *Element
     if cli.argc != 2 {
         cli.ErrorResponse(wrongArgs, "get")
         return nil
     }
     key := string(cli.argv[1])
     db := p.ctx.s.db[cli.selectDb]
     val = db.LookupKey(key)
     if val == nil {
         cli.Write(wrongKey)
         return nil
     }
     if val.Type != JON_STRING {
         cli.Write(wrongType)
         return nil
     }
     valStr, _ := val.Value.(string)
     valLen := len(valStr)
     respStr := fmt.Sprintf("$%d\r\n%s\r\n", valLen, valStr)
     cli.Write(respStr)
     return nil
}
*/
