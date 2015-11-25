package server

import (
    "fmt"
    "strings"
    //"time"
    //"strconv"
)

func (s *Server)Del(cli *Client) error {
    if cli.argc <= 1 {
        cli.ErrorResponse(wrongArgs, "del")
        return nil
    }
    db := s.db[cli.selectDb]
    del_num := 0
    for i := 1; i < int(cli.argc); i ++ {
        key_str := string(cli.argv[i])
        if ok := db.DeleteKey(key_str); ok {
            del_num ++
        }
    }
    resp := fmt.Sprintf(":%d\r\n", del_num)
    cli.Write(resp)
    return nil
}

func (s *Server)Keys(cli *Client) error {
    var resp string
    if cli.argc != 2 {
        cli.ErrorResponse(wrongArgs, "keys")
        return nil
    }
    key_str := string(cli.argv[1])
    db := s.db[cli.selectDb]
    key_num := 0
    if len(key_str) == 1 && key_str == "*" {
        respKeys := db.Keys()
        key_num = len(respKeys)
        resp = fmt.Sprintf("*%d\r\n", key_num)
        for _, k := range respKeys {
            resp = fmt.Sprintf("%s$%d\r\n%s\r\n", resp, len(k), k)
        }
        cli.Write(resp)

    } else {
        hasWild := strings.Contains(key_str, "*")
        if hasWild == false {
            if db.Haskey(key_str) {
                resp = fmt.Sprintf("*1\r\n$%d\r\n%s\r\n", len(key_str), key_str)
            } else {
                resp = fmt.Sprintf("*0\r\n")
            }
            cli.Write(resp)
        } else {
            
        }
    }
    return nil
}

func (s *Server)Type(cli *Client) error {
    if cli.argc != 2 {
        cli.ErrorResponse(wrongArgs, "type")
        return nil
    }
    var resp string
    db := s.db[cli.selectDb]
    key := string(cli.argv[1])
    db.RLock()
    val := db.LookupKey(key)
    if val == nil {
        resp = "none"
    } else {
        switch val.Type {
        case JON_STRING:
            resp = "string"
        case JON_HASH:
            resp = "hash"
        case JON_LIST:
            resp = "list"
        case JON_SET:
            resp = "set"
        case JON_SORTSET:
            resp = "zset"
        default:
            resp = "none"
        }
    }
    db.RUnlock()
    resp = fmt.Sprintf("+%s\r\n", resp)
    err := cli.Write(resp)
    return err
}
