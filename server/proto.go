package server

import (
    "io"
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
    var cmdNum, i int
    var line []byte

    srvExitChan = p.ctx.s.exitChan
    client := NewClient(conn)
    client.JonDb = p.ctx.s.db[client.selectDb]
    r := client.reader
    w := client.writer
    cliEixtChan = client.exitChan
    for {
        select {
        case <- cliExitChan:
            goto end
        case <- srvExitChan:
            goto end
        default:
        }
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
            if len(line) <= 4 || line[0] != '$' {
                io.ReadFull(r, dump)
                client.ErrorResponse("ERR unknown command '%s'", line)
                break
            }
            nextNum, err := strconv.Atoi(string(line[1:len(line)-2]))
            if err != nil {
                goto ERR
            }
            cmd := make([]byte, nextNum+2)
            io.ReadFull(r, cmd)
            client.argv = append(client.argv, cmd[0:len(cmd)-2])
        }
        err := p.processCommand(client)
        if err != nil {
            goto ERR
        }
ERR:
        io.ReadFull(r, dump)
        client.ErrorResponse("ERR unknown command '%s'", line)
        continue
    }
end:
    client.Exit()
}

func (p *JonProtocol) processCommand(cli *Client) {

}
