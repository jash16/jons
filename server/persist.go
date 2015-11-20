package server

func (s *Server)Bgsave(cli *Client) error {
    if cli.argc != 1 {
        cli.ErrorResponse(wrongArgs, "bgsave")
        return nil
    }
    var db []*JonDb

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
        db = append(db, nd)
        d.RUnlock()
    }
    return s.rdbSave(db)
}
