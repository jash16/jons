package server

func (s *Server)Bgsave(cli *Client) error {
    if cli.argc != 1 {
        cli.ErrorResponse(wrongArgs, "bgsave")
        return nil
    }
    var db []*JonDb

    for d := range s.db {
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
    now := time.Now()
    nowms := now.Unix() * 1000 + now.NanoSecond() / 100000

    for d = range db {
        for key, val := range d.Dict {
            expTime = -1
            if exp, ok := d.Expires[key]; ok {
                 expTime := exp.Value.(int64)
            }
            s.RdbSaveKeyValPair(key, val, expTime, nowms)
        }
    }
    return nil
}
