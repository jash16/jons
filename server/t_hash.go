package server

import (
    "fmt"
)

func (s *Server) Hset(cli *Client) error {
    var resp string
    if cli.argc != 4 {
        cli.ErrorResponse(wrongArgs, "hset")
        return nil
    }
    key_str := string(cli.argv[1])
    val_key := string(cli.argv[2])
    val_val := string(cli.argv[3])
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key_str)
    if ele == nil {
        val := make(map[string]string)
        val[val_key] = val_val
        ele := NewElement(JON_HASH, val)
        db.SetKey(key_str, ele)
        resp = fmt.Sprintf(":1\r\n")
    } else if ele.Type != JON_HASH {
        resp = wrongType
    } else {
        val := ele.Value.(map[string]string)
        if _, ok := val[val_key]; ok {
            resp = fmt.Sprintf(":0\r\n")
        } else {
            resp = fmt.Sprintf(":1\r\n")
        }
        val[val_key] = val_val
        ele.Value = val
        db.SetKey(key_str, ele)
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Hget(cli *Client) error {
    if cli.argc != 3 {
        cli.ErrorResponse(wrongArgs, "hget")
        return nil
    }
    var resp string
    key_str := string(cli.argv[1])
    val_key := string(cli.argv[2])
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key_str)
    if ele == nil {
        resp = wrongKey
    } else if ele.Type != JON_HASH {
        resp = wrongType
    } else {
        val := ele.Value.(map[string]string)
        //var val_val string
        if val_val, ok := val[val_key]; ok {
            resp = fmt.Sprintf("$%d\r\n%s\r\n", len(val_val), val_val)
        } else {
            resp = wrongKey
        }
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Hgetall(cli *Client) error {
    return nil
}

func (s *Server) Hkeys(cli *Client) error {
    return nil
}

func (s *Server) Hlen(cli *Client) error {
    var resp string
    if cli.argc != 2 {
        cli.ErrorResponse(wrongArgs, "hlen")
        return nil
    }
    key_str := string(cli.argv[1])
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key_str)
    if ele == nil {
        resp = zeroKey
    } else if ele.Type != JON_HASH {
        resp = wrongType
    } else {
        val := ele.Value.(map[string]string)
        length := len(val)
        resp = fmt.Sprintf(":%d\r\n", length)
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Hdel(cli *Client) error {
    if cli.argc <= 2 {
        cli.ErrorResponse(wrongArgs, "hdel")
        return nil
    }
    var resp string
    db := s.db[cli.selectDb]
    key := string(cli.argv[1])
    ele := db.LookupKey(key)
    if ele == nil {
        resp = zeroKey
    } else if ele.Type != JON_HASH {
        resp = wrongKey
    } else {
        delNum := 0
        var val map[string]string
        for i := 2; i < int(cli.argc); i ++ {
            field := string(cli.argv[i])
            val = ele.Value.(map[string]string)
            if _, ok := val[field]; ok {
                delNum ++
                delete(val, field)
            }
        }
        if delNum > 0 {
            ele.Value = val
            db.SetKey(key, ele)
        }
        resp = fmt.Sprintf(":%d\r\n", delNum)
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Hexists(cli *Client) error {
    return nil
}

func (s *Server) Hmset(cli *Client) error {
    return nil
}

func (s *Server) Hmget(cli *Client) error {
    return nil
}

func (s *Server) Hsetnx(cli *Client) error {
    return nil
}
