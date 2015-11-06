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
    return &Element {
        Type: typ,
        Encode: JON_ENCODING_RAW,
        Value: value,
    }
}

func (e *Element) Copy() *Element {
    switch e.Value.(type) {
    case string:
    case map[string]string: //use for hash type
    case map[string]byte:   //use for set and sorted set
    case [][]byte: //use for list
    }
}

type DB struct {
    DataMap map[Key]Element
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
