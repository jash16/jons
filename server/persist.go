package server
import(
    "sync/atomic"
    "time"
)
type Persist interface {
    Save(db []*JonDb) error
    Load(file string) error
}

func (s *Server)dbCopy() []*JonDb {
    var dbs []*JonDb

    for _, d := range s.db {
        d.RLock()
        dict := d.Dict.Copy()
        expire := d.Expires.Copy()
        nd := &JonDb {
            Dict: dict,
            Expires: expire,
        }
        dbs = append(dbs, nd)
        d.RUnlock()
    }
    return dbs
}

func (s *Server)Bgsave(cli *Client) error {
    if cli.argc != 1 {
        cli.ErrorResponse(wrongArgs, "bgsave")
        return nil
    }
    var dbs []*JonDb
    var inRdb bool
    s.Lock()
    if s.rdbFlag {
        inRdb = true
    } else {
        s.rdbFlag = true
    }
    s.Unlock()
    if inRdb {
         err := cli.Write("-ERR already in bgsave\r\n")
         return err
    }
/*
    for _, d := range s.db {
        d.RLock()
        dict := d.Dict.Copy()
        expire := d.Expires.Copy()
        nd := &JonDb {
            Dict: dict,
            Expires: expire,
        }
        dbs = append(dbs, nd)
        d.RUnlock()
    }
    */
    dbs = s.dbCopy()
    err := cli.Write("+Background saving started\r\n")
    time.Sleep(100 * time.Second)
    go s.rdbSave(dbs)
    return err
}

func (s *Server) Save(cli *Client) error {
    if cli.argc != 1 {
        cli.ErrorResponse(wrongArgs, "bgsave")
        return nil
    }
    var dbs []*JonDb
    var resp string
    var inRdb bool
    s.Lock()
    if s.rdbFlag {
    //    s.Unlock()
        inRdb = true
    } else {
        s.rdbFlag = true
    }
    s.Unlock()
    if inRdb {
        err := cli.Write("-ERR already in bgsave\r\n")
        return err
    }
    dbs = s.dbCopy()
    err := s.rdbSave(dbs)
    if err != nil {
        resp = "-ERR\r\n"
    } else {
        resp = "+OK\r\n"
    }
    err = cli.Write(resp)
    return err
}

func (s *Server) rdbSave(dbs []*JonDb) error {
    if s.p != nil {
        return s.p.Save(dbs)
    }
    return nil
}

func (s *Server) Bgrewriteaof(cli *Client) error {
    resp := "+Background append only file rewriting started\r\n"
    if atomic.CompareAndSwapInt32(&s.aofFlag, 0, 1) {
        cli.Write(resp)
        db := s.dbCopy()
        err := s.aof.bgRewrite(db)
        if err != nil {
            s.logf("aof bgRewrite err - %s", err)
        }
    }
    return nil
}
