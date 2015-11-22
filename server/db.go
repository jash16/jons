package server

import (
    "time"
    "sync"
//    "container/list"
)

const (
    JON_STRING int32 = iota
    JON_LIST
    JON_HASH
    JON_SET
    JON_SORTSET
    JON_INT64

    JON_ENCODING_RAW
    JON_ENCODING_INT
    JON_ENCODING_HT
    JON_ENCODING_ZIPMAP
    JON_ENCODING_LINKEDLIST
    JON_ENCODING_ZIPLIST
    JON_ENCODING_INTSET
    JON_ENCODING_SKIPLIST

    JON_KEY_NOTEXIST
)

type Key struct {
    //Type int32
    //Ref int32
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
    case int64:
    default:
        return nil
    }

    return &Element {
        Type: typ,
        Encode: JON_ENCODING_RAW,
        Value: value,
        Ref: 1,
    }
}

func (e *Element) Copy() *Element {
    var val2 interface{}
    switch e.Value.(type) {
    case string:
        val2 = e.Value
    case int64:
        val2 = e.Value // for expire key
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
    DataMap map[string]*Element
    sync.RWMutex
}

type JonDb struct {
    Dict *DB
    Expires *DB
    Blocks *DB
    Ready *DB
    Watch *DB
    sync.RWMutex
}

func NewDB() *DB {
    return &DB{
        DataMap: make(map[string]*Element),
    }
}

func (d *DB) Copy() *DB {
    d2 := NewDB()
    for key, ele := range d.DataMap {
        k := key
        e := ele.Copy()
        d2.DataMap[k] = e
    }
    return d2
}

func (d *DB) Size() int {
    return len(d.DataMap)
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

func (d *JonDb) SetKey(K string, V *Element) {
    dict := d.Dict
//    dict.Lock()
//    defer dict.Unlock()
    dict.DataMap[K] = V
}

func (d *JonDb) SetExpire(K string, V *Element) {
    dict := d.Expires
    dict.DataMap[K] = V
}

func (d *JonDb) LookupKey(K string) *Element {
    d.ExpireKey(K)
    var ele *Element
    var ok bool
    dict := d.Dict
//    dict.RLock()
//    defer dict.RUnlock()
    if ele, ok = dict.DataMap[K]; ok {
        return ele
    }
    return nil
}

func (d *JonDb) LookupKeyType(K string) int32 {
    ele := d.LookupKey(K)
    if ele != nil {
        return ele.Type
    }
    return JON_KEY_NOTEXIST
}

func (d *JonDb) DeleteKey(K string) bool {
    expdict := d.Expires
    dict := d.Dict

//    expdict.Lock()
    delete(expdict.DataMap, K)
//    expdict.Unlock()

//    dict.Lock()
    if _, ok := dict.DataMap[K]; !ok {
//        dict.Unlock()
        return false
    }
    delete(dict.DataMap, K)
//    dict.Unlock()
    return true
}

func (d *JonDb) ExpireKey(K string) bool {
    expdict := d.Expires
    var expire int64
    var ele *Element
    var ok bool
    now_time  := time.Now()
    now_ms := int64(now_time.Nanosecond() / 100000)
    now_ms += now_time.Unix() * 1000
//    expdict.Lock()
    //defer expdict.Unlock()
    if ele, ok = expdict.DataMap[K]; !ok {
        //expdict.Unlock()
        return false
    }
    expire = ele.Value.(int64)
    println(expire, now_ms)
    if expire >= now_ms {
//        expdict.Unlock()
        return false
    }
    delete(expdict.DataMap, K)
//    expdict.Unlock()

    dict := d.Dict
//    dict.Lock()
    delete(dict.DataMap, K)
//    dict.Unlock()
    return true
}

func (d *JonDb) Keys() []string {
    dict := d.Dict
//    dict.Lock()
//    defer dict.Unlock()
    var resp []string
    for k, _ := range (dict.DataMap) {
        resp = append(resp, k)
    }
    return resp
}

func (d *JonDb) Haskey(K string) bool {
    d.ExpireKey(K)
    dict := d.Dict
//    dict.Lock()
//    defer dict.Unlock()
    if _, ok := dict.DataMap[K]; ok {
        return true
    }
    return false
}
