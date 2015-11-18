package server

import (
    "strconv"
    "fmt"
    "time"
)

func (s *Server) Set(cli *Client) error {
    if (cli.argc < 3) {
        cli.ErrorResponse(wrongArgs, "set")
        return nil
    }
    var expireFlag bool
    var existSetFlag bool
    var noExistSetFlag bool
    var expireTime int64 //unit ms
    key_str := string(cli.argv[1])
    val_str := string(cli.argv[2])
    s.logf("receive command: %s %s %s", cli.argv[0], cli.argv[1], cli.argv[2])
    if cli.argc > 3 {
        for i := 3; i < int(cli.argc); i ++ {
            arg := string(cli.argv[i])
            if arg == "nx" || arg == "NX" {
                noExistSetFlag = true
            } else if arg == "xx" || arg == "XX" {
                existSetFlag = true
            } else if arg == "px" || arg == "PX" { //ms
                if i == int(cli.argc - 1) {
                    cli.ErrorResponse(wrongSyntax)
                    return nil
                }
                expTime, err := strconv.Atoi(string(cli.argv[i+1]))
                if err != nil {
                    cli.ErrorResponse(wrongArgType)
                    return nil
                }
                i += 1
                expireTime += int64(expTime)
                expireFlag = true
            } else if arg == "ex" || arg == "EX" {
                if i == int(cli.argc - 1) {
                    cli.ErrorResponse(wrongSyntax)
                    return nil
                }
                expTime, err := strconv.Atoi(string(cli.argv[i+1]))
                if err != nil {
                    cli.ErrorResponse(wrongArgType)
                    return nil
                }
                i += 1
                expireTime += int64(expTime * 1000)
                expireFlag = true
            }
        }
    }
    K := key_str
    V := NewElement(JON_STRING, val_str)
    db := s.db[cli.selectDb]
    typ := db.LookupKeyType(K)
    if typ != JON_KEY_NOTEXIST && typ != JON_STRING {
        cli.Write(wrongType)
        return nil
    }
    if (typ == JON_KEY_NOTEXIST && existSetFlag) || (typ != JON_KEY_NOTEXIST && noExistSetFlag) {
        cli.Write(zeroKey)
        return nil
    }
    db.SetKey(K, V)

    if expireFlag == true {
        t := expireTime
        curTimes := time.Now()
        curTimems := int64(curTimes.Nanosecond() / 100000)
        expireTime += curTimes.Unix()*1000 + curTimems
        s.logf("expireTime: %d(ms), expired time: %d, now time: %d", t, expireTime, curTimes.Unix()*1000 + curTimems)
        exp := NewElement(JON_INT64, expireTime)
        db.SetExpire(K, exp)
    }

    cli.Write(ok)
    return nil
}

func (s *Server) Select(cli *Client) error {
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

func (s *Server) genericGet(key string) *Element {
    return nil
}

func (s *Server) Get(cli *Client) error {
     var val *Element
     if cli.argc != 2 {
         cli.ErrorResponse(wrongArgs, "get")
         return nil
     }
     key := string(cli.argv[1])
     db := s.db[cli.selectDb]
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

func (s *Server)Mget(cli *Client) error {
    if cli.argc < 2 {
        cli.ErrorResponse(wrongArgs, "mget")
        return nil
    }
    db := s.db[cli.selectDb]

    var resp string
    resp += fmt.Sprintf("*%d\r\n", cli.argc - 1)
    for i := 1; i < int(cli.argc); i ++ {
        key := string(cli.argv[1])
        //db := s.db[cli.selectDb]
        val := db.LookupKey(key)
        if val == nil || val.Type != JON_STRING {
            resp += fmt.Sprintf("$-1\r\n")
        } else {
            valStr := val.Value.(string)
            resp += fmt.Sprintf("$%d\r\n", len(valStr))
            resp += fmt.Sprintf("%s\r\n", valStr)
        }
    }
    cli.Write(resp)
    return nil
}

func (s *Server)Mset(cli *Client) error {
    if (cli.argc - 1) % 2 != 0 {
        cli.ErrorResponse(wrongArgs, "mset")
        return nil
    }

    db := s.db[cli.selectDb]
    for i := 1; i < int(cli.argc); i += 2 {
        key_str := string(cli.argv[i])
        val_str := string(cli.argv[i+1])
        val := NewElement(JON_STRING, val_str)
        db.SetKey(key_str, val)
    }
    cli.Write(ok)
    return nil
}

func (s *Server)Append(cli *Client) error {

    return nil
}

func (s *Server)Bitcount(cli *Client) error {
    return nil
}

func (s *Server) Decr(cli *Client) error {
    return nil
}

func (s *Server) Decrby(cli *Client) error {
    return nil
}

func (s *Server) Getbit(cli *Client) error {
    return nil
}

func (s *Server) Getrange(cli *Client) error {
    return nil
}

func (s *Server) Getset(cli *Client) error {
    if cli.argc != 3 {
        cli.ErrorResponse(wrongArgs, "getset")
        return nil
    }
    key_str := string(cli.argv[1])
    val_str := string(cli.argv[2])
    db := s.db[cli.selectDb]
    ele := db.LookupKey(key_str)
    val := NewElement(JON_STRING, val_str)
    if ele == nil {
        db.SetKey(key_str, val)
        cli.Write(wrongKey)
    } else if ele.Type != JON_STRING {
        cli.Write(wrongType)
    } else {
        db.SetKey(key_str, val)
        val_str := val.Value.(string)
        resp := fmt.Sprintf("$%d\r\n%s\r\n", len(val_str), val_str)
        cli.Write(resp)
    }
    return nil
}

func (s *Server) Incr(cli *Client) error {
    return nil
}

func (s *Server) Incrby(cli *Client) error {
    return nil
}

func (s *Server) Incrbyfloat(cli *Client) error {
    return nil
}

func (s *Server) Strlen(cli *Client) error {
    if cli.argc != 2 {
        cli.ErrorResponse(wrongArgs, "strlen")
        return nil
    }
    db := s.db[cli.selectDb]
    key_str := string(cli.argv[1])
    ele := db.LookupKey(key_str)
    if ele == nil {
        cli.Write(zeroKey)
        return nil
    } else if ele.Type != JON_STRING {
        cli.Write(wrongType)
        return nil
    }
    val_str := ele.Value.(string)
    length := len(val_str)
    resp := fmt.Sprintf(":%d\r\n", length)
    cli.Write(resp)
    return nil
}
