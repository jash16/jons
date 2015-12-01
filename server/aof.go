package server

import (
    "os"
    "fmt"
    "common"
)

type aof struct {
    aofSelectDb int32
    aofHandler *os.File
}

func (o *aof) appendCmdSToFile(cmds []dirtyCmd) error {
    var fname string
    //fname = fmt.Sprintf("temp-%d.aof", os.Getpid())
    fname = fmt.Sprintf("append.aof")
    f, err := os.OpenFile(fname, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        return err
    }
    o.aofHandler = f
    var cmdString string

    for _, cmd := range cmds {
        if o.aofSelectDb != cmd.selectDb {
            s := fmt.Sprintf("%d", cmd.selectDb)
            cmdString = fmt.Sprintf("*2\r\n$6\r\nselect\r\n$%d\r\n%d\r\n", len(s), cmd.selectDb)
            o.aofSelectDb = cmd.selectDb
        }
        argc := cmd.argc
        argvs := cmd.argv
        cmdString = fmt.Sprintf("%s*%d\r\n", cmdString, argc)
        for _, argv := range argvs {
            cmdString = fmt.Sprintf("%s$%d\r\n%s\r\n", cmdString, len(argv), string(argv))
        }
    }
    fmt.Printf("%s\n", cmdString)
    err = o.Write(cmdString)
    if err != nil {
        return nil
    }
    o.Flush()
    os.Rename(fname, "append.aof")
    return nil
}

func (o *aof) Write(cmd string) error {
    _, err := o.aofHandler.WriteString(cmd)
    if err != nil {
        return err
    }
    return nil
}

func (o *aof) Flush() error {
    return o.aofHandler.Sync()
}

func (o *aof) bgRewrite(dbs []*JonDb) error {
    var err error
    var fname string
    var tnow int64

    tnow = common.GetMsTime()
    fname = fmt.Sprintf("temp-%d.aof", os.Getpid())
    f, err := os.OpenFile(fname, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        return err
    }
    o.aofHandler = f

    for i, db := range dbs {
        if db.Dict.Size() <= 0 {
            continue
        }
        is := fmt.Sprintf("%d", i)
        selectCmd := fmt.Sprintf("*2\r\n$6\r\nselect\r\n$%d\r\n%d\r\n", len(is), i)
        o.Write(selectCmd)
        var ex bool = false
        var extime int64
        var cmd string
        for key, ele := range db.Dict.DataMap {
            exp, ok := db.Expires.DataMap[key]
            if ok {
                 extime = exp.Value.(int64)
                 if extime <= tnow {
                     continue
                 }
                 ex = true
            } else {
                ex = false
            }
            switch ele.Type {
            case JON_STRING:
                val := ele.Value.(string)
                cmd = fmt.Sprintf("*3\r\n$3\r\nset\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
                                       len(key), key, len(val), val)
            case JON_LIST:
                val := ele.Value.([][]byte)
                valLen := len(val)
                cmd = fmt.Sprintf("*%d\r\n$5\r\nlpush\r\n$%d\r\n%s\r\n", 2+valLen, len(key), key)
                for _, v := range val {
                    cmd = fmt.Sprintf("%s$%d\r\n%s\r\n", cmd, len(v), v)
                }
            case JON_HASH:
                val := ele.Value.(map[string]string)
                mapLen := len(val)
                cmd = fmt.Sprintf("*%d\r\n$5\r\nhmset\r\n$%d\r\n%s\r\n", 2 + 2 * mapLen, len(key), key)
                for k, v := range val {
                    cmd = fmt.Sprintf("%s$%d\r\n%s\r\n$%d\r\n%s\r\n", cmd, len(k), k, len(v), v)
                }

            case JON_SET:
            case JON_SORTSET:
            }
            o.Write(cmd)
            if ex {
                expCmd := fmt.Sprintf("*3\r\n$9\r\npexpireat\r\n$%d\r\n%s\r\n$%d\r\n%d\r\n",
                                       len(key), key, len(fmt.Sprintf("%d", extime)), extime)
                o.Write(expCmd)
            }
        }
    }
    o.Flush()
    os.Rename(fname, "append.aof")
    return nil
}
