package server

import (
_    "fmt"
)

func (s *Server) Sadd(cli *Client) error {
    if cli.argc <= 2 {
         cli.ErrorResponse(wrongArgs, "sadd")
         return nil
    }

    //var resp string

    return nil
}

func (s *Server) Srem(cli *Client) error {

    return nil
}

func (s *Server) Sinter(cli *Client) error {

    return nil
}
