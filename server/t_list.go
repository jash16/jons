package server

import (
    "fmt"
    "strconv"
)

func (s *Server)Lpush(cli *Client) error {
    var val [][]byte
    var resp string
    if cli.argc <= 2 {
        cli.ErrorResponse(wrongArgs, "lpush")
        return nil
    }
    key_str := string(cli.argv[1])

    db := s.db[cli.selectDb]
    db.Lock()
    old_ele := db.LookupKey(key_str)
    if old_ele == nil {
        for i := 2; i < int(cli.argc); i ++ {
            val = append(val, cli.argv[i])
        }

        ele := NewElement(JON_LIST, val)
        db.SetKey(key_str, ele)
        resp = fmt.Sprintf(":%d\r\n", len(val))
        //cli.Write(resp)
    } else if old_ele.Type != JON_LIST {
        resp = wrongType
        //cli.Write(wrongType)
    } else {
        val_old := old_ele.Value.([][]byte)
        for i := 2; i < int(cli.argc); i ++ {
            val_old = append(val_old, cli.argv[i])
        }

        length := len(val_old)
        old_ele.Value = val_old
        db.SetKey(key_str, old_ele)
        resp = fmt.Sprintf(":%d\r\n", length)
        //cli.Write(resp)
    }
    db.Unlock()
    cli.Write(resp)
    return nil
}

func (s *Server) Lrange(cli *Client) error {
    if cli.argc != 4 {
        cli.ErrorResponse(wrongArgs, "lrange")
        return nil
    }
    var resp string
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
    db.RLock()
    ele := db.LookupKey(key)
    if ele == nil {
        resp = zeroLine
    } else if ele.Type != JON_LIST {
        resp = wrongType
    } else {
        val := ele.Value.([][]byte)
        val_num := len(val)
        if end == -1 {
            end = val_num -1
        } else if end > val_num {
            end = val_num
        }
        if start < -(val_num + 1) {
            start = 0
        } else if start < 0 && start >= -val_num {
            start += val_num
        }
        if start > end {
            resp = zeroLine
        } else {
            send_val_num := end - start + 1
            resp = fmt.Sprintf("*%d\r\n", send_val_num)
            for i := start; i <= end; i ++ {
                cur_val := val[i]
                cur_len := len(cur_val)
                resp = fmt.Sprintf("%s$%d\r\n%s\r\n", resp, cur_len, cur_val)
            }
        }
    }
    db.RUnlock()
    cli.Write(resp)
    return nil
}
