package server

import (
    "fmt"
)
type Pub struct {
    key string
    val string
}

type subExit struct {
    key string
    cli *Client
}

func (s *Server)Publish(cli *Client) error {
    if cli.argc != 3 {
        cli.ErrorResponse(wrongArgs, "publish")
        return nil
    }
    var resp string
    //var clis []*Client
    var succ int = 0
    pubkey := string(cli.argv[1])
    pubval := string(cli.argv[2])

    s.subLock.RLock()
    if clis, ok := s.subMap[pubkey]; ! ok {
        resp = zeroKey
    } else {
        var nclis []*Client
        for _, c := range clis {
            c.Lock()
            if c.subChan != nil {
                succ += 1
                pubdata := Pub {
                    key: pubkey,
                    val: pubval,
                }
                c.subChan <- pubdata
                nclis = append(nclis, c)
            }
            c.Unlock()
        }
        resp = fmt.Sprintf(":%d\r\n", succ)
        if len(nclis) == 0 {
            delete(s.subMap, pubkey)
        } else {
            s.subMap[pubkey] = nclis
        }
    }
    s.subLock.RUnlock()
    cli.Write(resp)
    return nil
}

func (s *Server)Subscribe(cli *Client) error {
    if cli.argc < 2 {
        cli.ErrorResponse(wrongArgs, "subscribe")
        return nil
    }
    for i := 1; i < int(cli.argc); i ++ {
        subkey := string(cli.argv[i])
        s.AddSubClient(subkey, cli)
        cli.subKeys = append(cli.subKeys, subkey)
    }
    for {
        select {
        case pubdata := <- cli.subChan:
            resp := fmt.Sprintf("*3\r\n$7\r\nmessage\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
                                len(pubdata.key), pubdata.key, len(pubdata.val), pubdata.val)
            err := cli.Write(resp)
            if err != nil {
                return err
            }
        case <- s.exitChan:
            goto end
        case <- cli.exitChan:
            goto end
        }
    }
end:
    return nil
}
