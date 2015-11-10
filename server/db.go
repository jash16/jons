package server

import (
    "sync"
//    "container/list"
)

const (
    JON_STRING iota
    JON_LIST
    JON_HASH
    JON_SET
    JON_SORTSET

    JON_ENCODING_RAW
    JON_ENCODING_INT
    JON_ENCODING_HT
    JON_ENCODING_ZIPMAP
    JON_ENCODING_LINKEDLIST
    JON_ENCODING_ZIPLIST
    JON_ENCODING_INTSET
    JON_ENCODING_SKIPLIST
)

type Key struct {
    Type int32
    Ref int32
    Value string
}

type Element struct {
    Type int32
    Encode int32
    Ref    int32
    Value interface{}
}

func NewElement(typ int32, value interface{}) *Element {

    switch value.(type) {
    case string:
    case map[string]string:
    case map[string]byte:
    case [][]byte:
    default:
        return nil
    }

    return &Element {
        Type: typ,
        Encode: JON_ENCODING_RAW,
        Value: value,
        Ref: 0,
    }
}

func (e *Element) Copy() *Element {
    var val2 interface{}
    switch e.Value.(type) {
    case string:
        val2 = e.Value

    case map[string]string: { //use for hash type
        val := make(map[string]string)
        v, _ := e.Value.(map[string]string)
        for key, value := range v {
            val[key] = value
        }
        val2 = val
    }
    case map[string]byte:{   //use for set and sorted set
        val := make(map[string]byte)
        v, _ := e.Value.(map[string]byte)
        for key, value := range v {
            val[key] = value
        }
        val2 = val
    }
    case [][]byte:{ //use for list
        v, _ := e.Value.([][]byte)
        val := make([][]byte, len(v))
        for i, _ := range v {
            val[i] = make([]byte, len(v[i]))
            copy(val[i], v[i])
        }
            val2 = val
    }
    default:
        return nil
    }

    return &Element {
        Type: e.Type,
        Ref: e.Ref,
        Encode: e.Encode,
        Value: val2,
    }
}

type DB struct {
    DataMap map[Key]*Element
    sync.RWMutex
}

type JonDb struct {
    Dict *DB
    Expires *DB
    Blocks *DB
    Ready *DB
    Watch *DB
    sync.Mutex
}

func NewDB() *DB {
    return &DB{
        DataMap: make(map[Key]Element),
    }
}

func (d *DB) Copy() *DB {
    d2 := NewDB()
    d.RLock()
    defer d.RUnlock()
    for key, ele := range d.DataMap {
        k := key,
        e := ele.Copy()
        d2.DataMap[k] = e
    }
}

func NewJonDb() *JonDb {
    return &JonDb {
        Dict: NewDB(),
        Expires: NewDB(),
        Blocks: NewDB(),
        Ready: NewDB(),
        Watch: NewDB(),
    }
}

func (d *JobDb) Persist() {

}
