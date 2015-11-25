package server

import (
    "time"
    "fmt"
    "io"
    "os"
    "net"
    "strconv"
    "encoding/binary"
    "bytes"
)

type rdb struct {
    rdbHandler *os.File
    slaves []net.Conn
    typ int
    ctx *context
}

const (
    RDBFILE int = 0
    RDBSLAVE int = 1
)
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
    REDIS_RDB_TYPE_HASH   uint8 = 2
    REDIS_RDB_TYPE_SET    uint8 = 3
    REDIS_RDB_TYPE_ZSET   uint8 = 4
    REDIS_RDB_TYPE_INT64  uint8 = 5
)

func (r *rdb) Init(typ int) {
    switch typ {
    case RDBFILE:
    case RDBSLAVE:
    }
}

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
    r.rdbHandler.Close()
    os.Rename(fname, "dump.rdb")
    return nil
ERROR:
    r.ctx.s.logf("save rdb get error - %s", err)
    return err
}

func (r *rdb) saveKeyValPair(key string, val *Element, expTime int64, nowTime int64) error {
    var err error
    if expTime != -1 {
        r.ctx.s.logf("exp is: %d, now is: %d", expTime, nowTime)
        if expTime <= nowTime { //has expire
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
        var num int
        vals = val.Value.([][]byte)
        num = len(vals)
        if err = r.saveLen(num); err != nil {
            return err
        }
        for _, valByte := range vals {
            if err = r.saveStr(string(valByte)); err != nil {
                return err
            }
        }
    case JON_HASH:
       var valMap map[string]string
       var num int
       valMap = val.Value.(map[string]string)
       num = len(valMap)
       if err = r.saveLen(num); err != nil {
           return err
       }
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

func (r *rdb) Load(file string)(error) {

    var now int64
    var err error
    var expTime int64 = -1
    var dbidx int
    var db *JonDb
    var key string
    var val *Element

    now = time.Now().Unix() * 1000 + int64(time.Now().Nanosecond()/100000)

    //dbs := r.ctx.s.db

    if f, err := os.Open(file); err != nil {
        return err
    } else {
        r.rdbHandler = f
    }
    var magic []byte
    if magic, err = r.readLen(9); err != nil {
        return err
    }
    if string(magic[0:5]) != "REDIS" {
        return fmt.Errorf("not redis rdb file")
    }
    var ver int
    if ver, err = strconv.Atoi(string(magic[5:9])); err != nil {
        return err
    }
    if uint8(ver) != REDIS_RDB_VERSION {
        return fmt.Errorf("rdb file version wrong")
    }
    var typ uint8
    for {
        if typ, err = r.loadType(); err != nil {
            fmt.Printf("load type error: %s\n", err)
            return err
        }
        if typ == REDIS_RDB_OPCODE_EXPIRETIME_MS {
            if expTime, err = r.loadInt64(); err != nil {
                fmt.Printf("load int64  error: %s\n", err)
                return err
            }
            if typ, err = r.loadType(); err != nil {
                fmt.Printf("load type error: %s\n", err)
                return err
            }
        }

        if typ == REDIS_RDB_OPCODE_EOF {
            break
        }
        if typ == REDIS_RDB_OPCODE_SELECTDB {
            if dbidx, err = r.loadLen(); err != nil {
                fmt.Printf("load dbidx error: %s\n", err)
                return err
            }
            if dbidx > r.ctx.s.Opts.DbNum {
                r.ctx.s.logf("load db index : %d bigger than db numer: %d", dbidx, r.ctx.s.Opts.DbNum)
                os.Exit(1)
            }
            db = r.ctx.s.db[dbidx]
            continue
        }
        if key, err = r.loadKey(); err != nil {
            return err
        }
        if val, err = r.loadVal(typ); err != nil {
            fmt.Printf("load val - %s\n", err)
            return err
        }
        if expTime != -1 && expTime < now {
            continue
        }
        db.SetKey(key, val)
        if expTime != -1 {
            ev := NewElement(JON_INT64, expTime)
            db.SetExpire(key, ev)
        }
    }
    return nil
}

func (r *rdb) loadType() (uint8, error) {
    var typ uint8
    data, err := r.readLen(1)
    if err != nil {
        return uint8(0), err
    }
    buf := bytes.NewReader(data)
    err = binary.Read(buf, binary.BigEndian, &typ)
    if err != nil {
        fmt.Printf("binary read error: %s\n", err)
        return uint8(0), err
    }
    return typ, nil
}

func (r *rdb) loadInt64() (int64, error) {
    var val64 int64
    var data []byte
    var err error
    if data, err = r.readLen(8); err != nil {
        return int64(0), err
    }
    buf := bytes.NewReader(data)
    err = binary.Read(buf, binary.BigEndian, &val64)
    if err != nil {
        return int64(0), err
    }
    return val64, nil
}

func (r *rdb) loadLen() (int, error) {
    var err error
    var typ uint8
    var typ2 uint8
    var length int32
    if typ, err = r.loadType(); err != nil {
        return 0, err
    }
    t := (typ & 0xC0) >> 6
    if t == REDIS_RDB_6BITLEN {
        return int(typ & 0x0000003F), nil
    } else if t == REDIS_RDB_14BITLEN {
        if typ2, err = r.loadType(); err != nil {
            fmt.Printf("load tpe error: %s\n", err)
            return 0, err
        }
        return int(((typ & 0x3F)<<8) | typ2), nil
    } else {
        var data []byte
        if data, err = r.readLen(4); err != nil {
            return 0, err
        }
        buf := bytes.NewReader(data)
        err = binary.Read(buf, binary.BigEndian, &length)
        if err != nil {
            fmt.Printf("binary read error %s\n", err)
            return 0, err
        }
        return int(length), nil
    }
    return 0, nil
}

func (r *rdb) loadKey() (string, error) {
    return r.loadStr()
}

func (r *rdb) loadStr() (string, error) {
    var length int
    var err error
    if length, err = r.loadLen(); err != nil {
        return "", err
    }
    buf := make([]byte, length)
    if _, err = io.ReadFull(r.rdbHandler, buf); err != nil {
        return "", err
    }
    return string(buf), nil
}

func (r *rdb) loadVal(typ uint8) (*Element, error) {
    var str string
    var err error
    switch typ{
    case REDIS_RDB_TYPE_STRING:
        if str, err = r.loadStr(); err != nil {
            return nil, err
        }
        return NewElement(JON_STRING, str), nil
    case REDIS_RDB_TYPE_LIST:
        var num int
        var vals [][]byte
        if num, err = r.loadLen(); err != nil {
            return nil, err
        }
        for i := 0 ; i < num; i ++ {
            if str, err = r.loadStr(); err != nil {
                return nil, err
            }
            vals = append(vals, []byte(str))
        }
        return NewElement(JON_LIST, vals), nil
    case REDIS_RDB_TYPE_HASH:
        var num int
        var key, val string
        var dataMap map[string]string
        dataMap = make(map[string]string)
        if num, err = r.loadLen(); err != nil {
            fmt.Printf("load map len: %s\n", err)
            return nil, err
        }
        for i := 0; i < num; i ++ {
            if key, err = r.loadStr(); err != nil {
                fmt.Printf("load map key %s\n", err)
                return nil, err
            }
            if val, err = r.loadStr(); err != nil {
                fmt.Printf("load map val %s\n", err)
                return nil, err
            }
            dataMap[key] = val
        }
        return NewElement(JON_HASH, dataMap), nil
    case REDIS_RDB_TYPE_SET:
    case REDIS_RDB_TYPE_ZSET:
    default:
        return nil, nil
    }
    return nil, nil
}

func (r *rdb) readLen(length int) ([]byte, error) {
    buf := make([]byte, length)
    if _, err := io.ReadFull(r.rdbHandler, buf); err != nil {
        return nil, err
    }
    return buf, nil
}
