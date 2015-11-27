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
    for _, cmd := range cmds {
        if o.aofSelectDb != cmd.selectDb {
            s := fmt.Sprintf("%s", cmd.selectDb)
            selectCmd := fmt.Sprintf("*2d\r\n$6\r\nselect\r\n$%d\r\n%d\r\n", len(s), cmd.selectDb)
        }
        
    }
}
