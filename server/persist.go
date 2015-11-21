package server

type Persist interface {
    Save(db []*JonDb) error
    Load(file string) (db []*JonDb, err error)
}

func (s *Server)Bgsave(cli *Client) error {
    if cli.argc != 1 {
        cli.ErrorResponse(wrongArgs, "bgsave")
        return nil
    }
    var dbs []*JonDb

    s.Lock()
    if s.rdbFlag {
        s.Unlock()
        return nil
    }

    s.rdbFlag = true
    s.Unlock()

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
    err := cli.Write("+Background saving started\r\n")
    go s.rdbSave(dbs)
    return err
}

func (s *Server) rdbSave(dbs []*JonDb) error {
    if s.p != nil {
        return s.p.Save(dbs)
    }
    return nil
}
