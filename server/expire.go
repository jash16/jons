package server

import (
    "strconv"
    "time"
    "common"
)

func (s *Server) Expire(cli *Client) error {
    if cli.argc != 3 {
        return cli.ErrorResponse(wrongArgs, "expire")
    }

    var resp  string
    key := string(cli.argv[1])
    expireTime, err := strconv.Atoi(string(cli.argv[2]))
    if err != nil {
        return cli.ErrorResponse(wrongArgType)
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
        return cli.ErrorResponse(wrongArgs, "pexpire")
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

func (s *Server) PexpireAt(cli *Client) error {
    if cli.argc != 3 {
        return cli.ErrorResponse(wrongArgs, "pexpireat")
    }

    var resp string
    key := string(cli.argv[1])
    expireTime, err := strconv.Atoi(string(cli.argv[2]))
    if err != nil {
        return cli.ErrorResponse(wrongArgType)
    }
    expired := int64(expireTime) * 100000
    nowMs := common.GetMsTime()
    db := s.db[cli.selectDb]
    db.Lock()
    val := db.LookupKey(key)
    if val == nil {
        resp = zeroKey
    } else {
        if nowMs >= expired {
            db.DeleteKey(key)
        } else {
            ele := NewElement(JON_INT64, expired)
            db.SetExpire(key, ele)
        }
        resp = oneKey
    }
    db.Unlock()
    return cli.Write(resp)
}

func (s *Server) ExpireAt(cli *Client) error {
    if cli.argc != 3 {
        return cli.ErrorResponse(wrongArgs, "pexpireat")
    }

    var resp string
    key := string(cli.argv[1])
    expireTime, err := strconv.Atoi(string(cli.argv[2]))
    if err != nil {
        return cli.ErrorResponse(wrongArgType)
    }
    expired := int64(expireTime)
    nowMs := common.GetMsTime()
    db := s.db[cli.selectDb]
    db.Lock()
    val := db.LookupKey(key)
    if val == nil {
        resp = zeroKey
    } else {
        if nowMs >= expired {
            db.DeleteKey(key)
        } else {
            ele := NewElement(JON_INT64, expired)
            db.SetExpire(key, ele)
        }
        resp = oneKey
    }
    db.Unlock()
    return cli.Write(resp)
}
