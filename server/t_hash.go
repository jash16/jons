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
    if cli.argc != 2 {
        cli.ErrorResponse(wrongArgs, "hgetall")
        return nil
    }
    var resp string
    key := string(cli.argv[1])
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key)
    if ele == nil {
        resp = zeroLine
    } else if ele.Type != JON_HASH {
        resp = wrongType
    } else {
        val := ele.Value.(map[string]string)
        resp = fmt.Sprintf("*%d\r\n", 2 * len(val))
        for k, f := range val {
            resp = fmt.Sprintf("%s$%d\r\n%s\r\n$%d\r\n%s\r\n", resp, len(k), k, len(f), f)
        }
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Hkeys(cli *Client) error {
    if cli.argc != 2 {
        cli.ErrorResponse(wrongArgs, "hkeys")
        return nil
    }
    var resp string
    db := s.db[cli.selectDb]
    key := string(cli.argv[1])
    ele := db.LookupKey(key)
    if ele == nil {
        resp = zeroLine
    } else if ele.Type != JON_HASH {
        resp = wrongType
    } else {
        val := ele.Value.(map[string]string)
        resp = fmt.Sprintf("*%d\r\n", len(val))
        for k, _ := range val {
            resp = fmt.Sprintf("%s$%d\r\n%s\r\n", resp, len(k), k)
        }
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Hvals(cli *Client) error {
    if cli.argc != 2 {
        cli.ErrorResponse(wrongArgs, "hvals")
        return nil
    }
    var resp string
    db := s.db[cli.selectDb]
    key := string(cli.argv[1])
    ele := db.LookupKey(key)
    if ele == nil {
        resp = zeroLine
    } else if ele.Type != JON_HASH {
        resp = wrongType
    } else {
        val := ele.Value.(map[string]string)
        resp = fmt.Sprintf("*%d\r\n", len(val))
        for _, v := range val {
            resp = fmt.Sprintf("%s$%d\r\n%s\r\n", resp, len(v), v)
        }
    }
    cli.Write(resp)
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
    if cli.argc != 3 {
        cli.ErrorResponse(wrongArgs, "hexists")
        return nil
    }
    var resp string
    key := string(cli.argv[1])
    field := string(cli.argv[2])
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key)
    if ele == nil {
        resp = zeroKey
    } else if ele.Type != JON_HASH {
        resp = wrongType
    } else {
        val := ele.Value.(map[string]string)
        if _, ok := val[field]; ok {
            resp = oneKey
        } else {
            resp = zeroKey
        }
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Hmset(cli *Client) error {
    if cli.argc <= 2  || cli.argc % 2 != 0 {
        cli.ErrorResponse(wrongArgs, "hmset")
        return nil
    }
    var resp string
    key := string(cli.argv[1])
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key)
    if ele == nil {
        val := make(map[string]string)
        for i := 2; i < int(cli.argc); i += 2 {
            fkey := string(cli.argv[i])
            fval := string(cli.argv[i+1])
            val[fkey] = fval
        }
        e := NewElement(JON_HASH, val)
        db.SetKey(key, e)
        resp = ok
    } else if ele.Type != JON_HASH {
        resp = wrongType
    } else {
        val := ele.Value.(map[string]string)
        for i := 2; i < int(cli.argc); i += 2 {
            fkey := string(cli.argv[i])
            fval := string(cli.argv[i+1])
            val[fval] = fkey
        }
        ele.Value = val
        db.SetKey(key, ele)
        resp = ok
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Hmget(cli *Client) error {
    if cli.argc <= 2 {
        cli.ErrorResponse(wrongArgs, "hmget")
        return nil
    }
    var resp string
    key := string(cli.argv[1])
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key)
    if ele == nil {
        resp = fmt.Sprintf("*1\r\n$-1\r\n")
    } else if ele.Type != JON_HASH {
        resp = wrongType
    } else {
        val := ele.Value.(map[string]string)
        resp = fmt.Sprintf("*%d\r\n", cli.argc - 2)
        for i := 2; i < int(cli.argc); i ++ {
            fkey := string(cli.argv[i])
            if fval, ok := val[fkey]; ok {
                resp = fmt.Sprintf("%s$%d\r\n%s\r\n", resp, len(fval), fval)
            } else {
                resp = fmt.Sprintf("%s$-1\r\n", resp)
            }
        }
    }
    cli.Write(resp)
    return nil
}

func (s *Server) Hsetnx(cli *Client) error {
    return nil
}
