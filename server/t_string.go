package server

import (
    "strconv"
    "fmt"
)

func (s *Server) Set(cli *Client) error {
    if (cli.argc != 3) {
        cli.ErrorResponse(wrongArgs, "set")
        return nil
    }
    key_str := string(cli.argv[1])
    val_str := string(cli.argv[2])
    s.logf("receive command: %s %s %s", cli.argv[0], cli.argv[1], cli.argv[2])
    K := key_str
    V := NewElement(JON_STRING, val_str)
    db := s.db[cli.selectDb]
    typ := db.LookupKeyType(K)
    if typ != JON_KEY_NOTEXIST && typ != JON_STRING {
        cli.Write(wrongType)
        return nil
    }
    db.SetKey(K, V)
    cli.Write(ok)
    return nil
}

func (s *Server) Select(cli *Client) error {
    if cli.argc != 2 {
        cli.ErrorResponse(wrongArgs, "select")
        return nil
    }
    dbIdx, _ := strconv.Atoi(string(cli.argv[1]))
    if dbIdx >= 16 || dbIdx < 0 {
        cli.ErrorResponse(wrongDbIdx)
        return nil
    }
    cli.selectDb = int32(dbIdx)
    cli.Write(ok)
    return nil
}

func (s *Server) Get(cli *Client) error {
     var val *Element
     if cli.argc != 2 {
         cli.ErrorResponse(wrongArgs, "get")
         return nil
     }
     key := string(cli.argv[1])
     db := s.db[cli.selectDb]
     val = db.LookupKey(key)
     if val == nil {
         cli.Write(wrongKey)
         return nil
     }
     if val.Type != JON_STRING {
         cli.Write(wrongType)
         return nil
     }
     valStr, _ := val.Value.(string)
     valLen := len(valStr)
     respStr := fmt.Sprintf("$%d\r\n%s\r\n", valLen, valStr)
     cli.Write(respStr)
     return nil
}
