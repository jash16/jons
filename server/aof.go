package server

import (
    "os"
    "fmt"
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

func (o *aof) bgRewrite(db []*JonDb) error {
    return nil
}
