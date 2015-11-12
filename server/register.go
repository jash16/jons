package server

func (s *Server) Register() {
// for string
    s.cmdMap["set"] = s.Set
    s.cmdMap["get"] = s.Get

//for db
    s.cmdMap["select"] = s.Get
}
