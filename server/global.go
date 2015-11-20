package server

import (
    "fmt"
    "strings"
    "time"
    "strconv"
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


func (s *Server) Expire(cli *Client) error {
    if cli.argc != 3 {
        cli.ErrorResponse(wrongArgs, "expire")
        return nil
    }

    var resp  string
    key := string(cli.argv[1])
    expireTime, err := strconv.Atoi(string(cli.argv[2]))
    if err != nil {
        cli.ErrorResponse(wrongArgType)
        return nil
    }
    expired := time.Now().Unix() * 1000 + int64(expireTime) * 1000
    s.logf("expire time: %d, expired: %d", expireTime, expired)
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key)
    if ele == nil {
        resp = zeroKey
    } else {
        val := NewElement(JON_INT64, expired)
        db.SetExpire(key, val)
        resp = oneKey
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Pexpire(cli *Client) error {
    if cli.argc != 3 {
        cli.ErrorResponse(wrongArgs, "pexpire")
        return nil
    }

    var resp  string
    key := string(cli.argv[1])
    expireTime, err := strconv.Atoi(string(cli.argv[2]))
    if err != nil {
        cli.ErrorResponse(wrongArgType)
        return nil
    }
    expired := time.Now().Unix() * 1000 + int64(expireTime)
    s.logf("expire time: %d, expired: %d", expireTime, expired)
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key)
    if ele == nil {
        resp = zeroKey
    } else {
        val := NewElement(JON_INT64, expired)
        db.SetExpire(key, val)
        resp = oneKey
    }
    cli.Write(resp)
    return nil
}
