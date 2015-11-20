package server

import (
    "time"
    "fmt"
    "os"
)

type Persist interface {
    Save(db []*JonDb) error
    Load(file string) (db []*JonDb, error)
}

type rdb struct {

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
    f, err := os.Open(fname)
    if err != nil {
        s.logf("open %s failed - %s", fname, err)
        return error
    }
    s.rdbHandler = f
    magic := fmt.Sprintf("REDIS%4d", REDIS_RDB_VERSION)
    s.rdbHandler.Write(magic)
    for i, d := range db {
        for key, val := range d.Dict.DataMap {
            expTime := -1
            if exp, ok := d.Expires.DataMap[key]; ok {
                 expTime := exp.Value.(int64)
            }
            s.RdbSaveKeyValPair(key, val, expTime, nowms)
        }
    }
    return nil
}

func (s *Server) RdbSaveKeyValPair(key string, val *Element, expTime int64, nowTime int64) {
    if expTime != -1 {
        if expTime >= nowTime {
            return
        }

    }
}

func (r *rdb) Load(file string) ([]*JonDb, error) {
    return nil, nil
}
