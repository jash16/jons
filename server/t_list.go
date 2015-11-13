package server

import (
    "fmt"
)

func (s *Server)Lpush(cli *Client) error {
    var val [][]byte
    if cli.argc <= 2 {
        cli.ErrorResponse(wrongArgs, "lpush")
        return nil
    }
    key_str := string(cli.argv[1])

    db := s.db[cli.selectDb]
    old_ele := db.LookupKey(key_str)
    if old_ele == nil {
        for i := 2; i < int(cli.argc); i ++ {
            val = append(val, cli.argv[i])
        }

        ele := NewElement(JON_LIST, val)
        db.SetKey(key_str, ele)
        resp := fmt.Sprintf(":%d\r\n", len(val))
        cli.Write(resp)
    } else if old_ele.Type != JON_LIST {
        cli.Write(wrongType)
    } else {
        val_old := old_ele.Value.([][]byte)
        for i := 2; i < int(cli.argc); i ++ {
            val_old = append(val_old, cli.argv[i])
        }

        length := len(val_old)
        old_ele.Value = val_old
        db.SetKey(key_str, old_ele)
        resp := fmt.Sprintf(":%d\r\n", length)
        cli.Write(resp)
    }
    return nil
}

func (s *Server) Lrange(cli *Client) error {
    if cli.argc != 4 {
        cli.ErrorResponse(wrongArgs, "lrange")
        return nil
    }

    key := string(cli.argv[1])
    start, err := strconv.Atoi(string(cli.argv[2]))
    if err != nil {
        cli.ErrorResponse("-ERR value is not an integer or out of range\r\n")
        return nil
    }
    end, err := strconv.Atoi(string(cli.argv[3]))
    if err != nil {
        cli.ErrorResponse("-ERR value is not an integer or out of range\r\n")
        return nil
    }

    db := s.db[cli.selectDb]
    ele := db.LookupKey(key)
    if ele == nil {
        resp := fmt.Sprintf("*0\r\n")
        cli.Write(resp)
    } else if ele.Type != JON_LIST {
        cli.Write(wrongType)
    } else {
        val := ele.Value.([][]byte)
        val_num := len(val)

    }
    return nil
}
