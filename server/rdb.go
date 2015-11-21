package server

import (
    "time"
    "fmt"
    "os"
)

type rdb struct {
    rdbHandler *os.File
}

const (
    REDIS_RDB_VERSION int = 6
    REDIS_RDB_OPCODE_EXPIRETIME_MS int = 252
    REDIS_RDB_OPCODE_EXPIRETIME int = 253
    REDIS_RDB_OPCODE_SELECTDB   int = 254
    REDIS_RDB_OPCODE_EOF        int = 255
)

func (r *rdb) Save(db []*JonDb) error {
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
        r.saveLen(i)
        for key, val := range d.Dict.DataMap {
            expTime := int64(-1)
            if exp, ok := d.Expires.DataMap[key]; ok {
                 expTime = exp.Value.(int64)
            }
            r.saveKeyValPair(key, val, expTime, nowms)
        }
    }
    os.Rename(fname, "dump.rdb")
    return nil
}

func (r *rdb) saveKeyValPair(key string, val *Element, expTime int64, nowTime int64) {
    if expTime != -1 {
        if expTime >= nowTime {
            return
        }
    }
}

func (r *rdb) saveLen(val int) {

}

func (r *rdb) Load(file string) ([]*JonDb, error) {
    return nil, nil
}
