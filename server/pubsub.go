package server

func (s *Server)Publish(cli *Client) error {
    if cli.argc != 3 {
        cli.ErrorResponse(wrongArgs, "publish")
        return nil
    }
    var resp string
    var clis []*Client
    var succ int = 0
    pubkey := string(cli.argv[1])
    pubval := string(cli.argv[2])

    s.subLock.RLock()
    if clis, ok := s.subMap[pubkey]; ! ok {
        resp = zeroKey
    } else {
        var nclis []*Client
        for idx, c := range clis {
            c.Lock()
            if c.subChan != nil {
                succ += 1
                c.subChan <- pubval
                nclis = append(nclis, c)
            }
            c.Unlock()
        }
        resp := fmt.Sprintf(":%d\r\n", succ)
        if len(nclis) == 0 {
            delete(s.subMap, pubkey)
        } else {
            s.subMap[pubkey] = nclis
        }
    }
    s.subLock.RUnlock()
    cli.Write(resp)
    return nil
}

func (s *Server)Subscribe(cli *Client) {
    if cli.argc <= 2 {
        cli.ErrorResponse(wrongArgs, "subscribe")
        return nil
    }
    for i := 1; i < int32(cli.argc); i ++ {
        subkey := string(cli.argv[i])
        s.AddSubClient(subkey, cli)
    }
    for {
        select {
        case data := <- c.subChan:
            cli.Write(data)
        }
    }
    return nil
}
