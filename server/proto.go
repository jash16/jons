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
            _, err = io.ReadFull(r, dump)
            if err != nil {
                goto end
            }
            err = client.ErrorResponse(wrongCommand, line)
            if err != nil {
                goto end
            }
            continue
        }
        cmdNum, err = strconv.Atoi(string(line[1:len(line)-2]))
        if err != nil {
            _, err = io.ReadFull(r, dump)
            if err != nil {
                goto end
            }
            err = client.ErrorResponse(wrongCommand, line)
            if err != nil {
                goto end
            }
            continue
        }
        for i := 0; i < cmdNum; i ++ {
            line, err = r.ReadSlice('\n')
            if err != nil {
                _, err = io.ReadFull(r, dump)
                if err != nil {
                    goto end
                }
                err = client.ErrorResponse(wrongCommand, line)
                if err != nil {
                    goto end
                }
                break
            }
            if len(line) < 4 || line[0] != '$' {
                _, err = io.ReadFull(r, dump)
                if err != nil {
                    goto end
                }
                err = client.ErrorResponse(wrongCommand, line)
                if err != nil {
                    goto end
                }
                break
            }
            nextNum, err = strconv.Atoi(string(line[1:len(line)-2]))
            if err != nil {
                goto ERR
            }
            cmd = make([]byte, nextNum+2)
            _, err = io.ReadFull(r, cmd)
            if err != nil {
                goto end
            }
            client.argv = append(client.argv, cmd[0:len(cmd)-2])
            client.argc ++;
        }
        err = p.processCommand(client)
        if err != nil {
            goto end
        } else {
            continue
        }
ERR:
        _, err = io.ReadFull(r, dump)
        if err != nil {
            goto end
        }
        err = client.ErrorResponse(wrongCommand, line)
        if err != nil {
            goto end
        }
    }
end:
    for _, key := range client.subKeys {
        println(key)
        subexit := subExit {
            key: key,
            cli: client,
        }
        p.ctx.s.subExitChan <- subexit
    }
    client.Exit()
    return err
}

func (p *JonProtocol) processCommand(cli *Client) error {

    var err error
    cmd := string(cli.argv[0])
    if cmdFunc, ok := p.ctx.s.cmdMap[cmd] ; ok {
        err = cmdFunc(cli)
    } else {
        err = cli.ErrorResponse(wrongCommand, cli.argv[0])
    }
    return err
}
