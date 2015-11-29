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
    fname = fmt.Sprintf("temp-%s.aof", os.Getpid())

    f, err := os.Create(fname)
    if err != nil {
        return nil
    }
    o.aofHandler = f
    for _, cmd := range cmds {
        if o.aofSelectDb != cmd.selectDb {
            s := fmt.Sprintf("%s", cmd.selectDb)
            selectCmd := fmt.Sprintf("*2d\r\n$6\r\nselect\r\n$%d\r\n%d\r\n", len(s), cmd.selectDb)
            err := o.Write(selectCmd)
            if err != nil {
                return nil
            }
            o.aofSelectDb = cmd.selectDb
        }
        argc := cmd.argc
        argvs := cmd.argv
        cmdString := fmt.Sprintf("*%d\r\n", argc)
        for _, argv := range argvs {
            cmdString = fmt.Sprintf("%s$%d\r\n%s\r\n", len(argv), argv)
        }
        err := o.Write(cmdString)
        if err != nil {
            return nil
        }
    }
    o.Flush()
    return nil
}

func (o *aof) Write(cmd string) error {
    _, err := o.aofHandler.Write([]byte(cmd))
    if err != nil {
        return err
    }
    return nil
}

func (o *aof) Flush() error {
    return o.aofHandler.Sync()
}
