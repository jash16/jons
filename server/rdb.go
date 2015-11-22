package server

import (
    "time"
    "fmt"
    "os"
    "encoding/binary"
    "bytes"
)

type rdb struct {
    rdbHandler *os.File
    ctx *context
}

const (
    REDIS_RDB_6BITLEN uint8 = 0
    REDIS_RDB_14BITLEN uint8 = 1
    REDIS_RDB_32BITLEN uint8 = 2
    REDIS_RDB_ENCVAL uint8 = 3

    REDIS_RDB_VERSION uint8 = 6
    REDIS_RDB_OPCODE_EXPIRETIME_MS uint8 = 252
    REDIS_RDB_OPCODE_EXPIRETIME uint8 = 253
    REDIS_RDB_OPCODE_SELECTDB   uint8 = 254
    REDIS_RDB_OPCODE_EOF        uint8 = 255

    REDIS_RDB_TYPE_STRING uint8 = 0
    REDIS_RDB_TYPE_LIST   uint8 = 1
    REDIS_RDB_TYPE_SET    uint8 = 2
    REDIS_RDB_TYPE_ZSET   uint8 = 3
    REDIS_RDB_TYPE_HASH   uint8 = 4
    REDIS_RDB_TYPE_INT64  uint8 = 5
)

func (r *rdb) Save(db []*JonDb) error {
    var err error
    var key string
    var val *Element

    now := time.Now()
    nowms := now.Unix() * 1000 + int64(now.Nanosecond() / 100000)

    fname := fmt.Sprintf("temp-%d.rdb", os.Getpid())
    //f, err := os.OpenFile(fname, os.O_WRONLY, 0666)
    f, err := os.Create(fname)
    if err != nil {
        fmt.Printf("open file: %s - %s", fname, err)
        return err
    }
    r.rdbHandler = f
    magic := fmt.Sprintf("REDIS%04d", REDIS_RDB_VERSION)
    r.rdbHandler.Write([]byte(magic))
    for i, d := range db {
        if d.Dict.Size() == 0 {
            continue
        }
        if err = r.saveType(REDIS_RDB_OPCODE_SELECTDB); err != nil {
            goto ERROR
        }

        if err = r.saveLen(i); err != nil {
            goto ERROR
        }
        for key, val = range d.Dict.DataMap {
            expTime := int64(-1)
            if exp, ok := d.Expires.DataMap[key]; ok {
                 expTime = exp.Value.(int64)
            }
            if err = r.saveKeyValPair(key, val, expTime, nowms); err != nil {
                goto ERROR
            }
        }
    }
    if err = r.saveType(REDIS_RDB_OPCODE_EOF); err != nil {
        goto ERROR
    }
    os.Rename(fname, "dump.rdb")
    return nil
ERROR:
    r.ctx.s.logf("save rdb get error - %s", err)
    return err
}

func (r *rdb) saveKeyValPair(key string, val *Element, expTime int64, nowTime int64) error {
    var err error
    if expTime != -1 {
        if expTime >= nowTime { //has expire
            return nil
        }
        if err = r.saveType(REDIS_RDB_OPCODE_EXPIRETIME_MS); err != nil {
            return err
        }
        if err = r.saveMillisecondTime(expTime); err != nil {
            return err
        }
    }
    if err = r.saveKeyType(val.Type); err != nil {
        return err
    }
    if err = r.saveKey(key); err != nil {
        return err
    }
    if err = r.saveVal(val); err != nil {
        return err
    }
    return err
}

func (r *rdb) saveMillisecondTime(exp int64) error {
    return r.saveInt64(exp)
}

func (r *rdb) saveLen(val int) error {
    var err error
    buf := new(bytes.Buffer)
    if val < (1 << 6) {
        var val8 uint8
        val8 = uint8(val & 0xFF)
        _, err = r.rdbHandler.Write([]byte{val8})
    } else if (val < (1<<14)) {
        var val8h, val8l uint8
        val8l = uint8((val>>8)&0xFF)|(REDIS_RDB_14BITLEN<<6)
        val8h = uint8(val&0xFF)
        _, err = r.rdbHandler.Write([]byte{val8l, val8h})
    } else {
        val8 := (REDIS_RDB_32BITLEN<<6)
        _, err := r.rdbHandler.Write([]byte{val8})
        if err != nil {
            goto ERROR
        }
        err = binary.Write(buf, binary.BigEndian, uint32(val))
        if err != nil {
            goto ERROR
        }
        _, err = r.rdbHandler.Write(buf.Bytes())
    }
ERROR:
    if err != nil {
        r.ctx.s.logf("saveLen error - %s\n", err)
    }
    return err
}

func (r *rdb) saveType(val uint8) error {
    _, err := r.rdbHandler.Write([]byte{val})
    return err
}

func (r *rdb) saveKeyType(val int32) error {
    val8 := uint8(val)
    _, err := r.rdbHandler.Write([]byte{val8})
    return err
}

func (r *rdb) saveKey(key string) error {
    return r.saveStr(key)
}

func (r *rdb) saveStr(key string) error {
    var err error
    if err = r.saveLen(len(key)); err != nil {
        return err
    }
    _, err = r.rdbHandler.Write([]byte(key))
    return err
}

func (r *rdb) saveInt64(val int64) error {
    buf := new(bytes.Buffer)
    if err := binary.Write(buf, binary.BigEndian, val); err != nil {
        return err
    }
    _, err := r.rdbHandler.Write(buf.Bytes())
    return err
}

func (r *rdb) saveVal(val *Element) error {
    var err error
    switch val.Type {
    case JON_STRING:
        valStr := val.Value.(string)
        return r.saveStr(valStr)
    case JON_LIST:
        var vals [][]byte
        vals = val.Value.([][]byte)
        for _, valByte := range vals {
            if err = r.saveStr(string(valByte)); err != nil {
                return err
            }
        }
    case JON_HASH:
       var valMap map[string]string
       valMap = val.Value.(map[string]string)
       for k, v := range valMap {
           if err = r.saveStr(k); err != nil {
               return err
           }
           if err = r.saveStr(v); err != nil {
               return err
           }
       }
    case JON_SET:
    case JON_SORTSET:
    }
    return err
}

func (r *rdb) Load(file string) ([]*JonDb, error) {
    return nil, nil
}
