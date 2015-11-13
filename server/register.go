package server

func (s *Server) Register() {
// for string
    s.cmdMap["set"] = s.Set
    s.cmdMap["get"] = s.Get
    s.cmdMap["mset"] = s.Mset
    s.cmdMap["mget"] = s.Mget
    s.cmdMap["strlen"] = s.Strlen

//for hash

//for list

//for set

//for zset

//for db
    s.cmdMap["select"] = s.Select
}
