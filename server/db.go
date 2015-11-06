package server

import (
    "sync"
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
    Value string
}

type Element struct {
    Type int32
    Encode int32
    Ref    int32
    Value interface{}
}

func NewElement(typ int32, value interface{}) *Element {
    return &Element {
        Type: typ,
        Encode: JON_ENCODING_RAW,
        Value:
    }
}

func (e *Element) Copy() *Element {

}
type DB struct {
    DataMap map[Element]Element
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
        DataMap: make(map[Element]Element),
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
