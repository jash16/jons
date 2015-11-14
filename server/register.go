package server

func (s *Server) Register() {
// for string
    s.cmdMap["set"] = s.Set
    s.cmdMap["get"] = s.Get
    s.cmdMap["mset"] = s.Mset
    s.cmdMap["mget"] = s.Mget
    s.cmdMap["strlen"] = s.Strlen
    s.cmdMap["getset"] = s.Getset

//for hash
    s.cmdMap["hset"] = s.Hset
    s.cmdMap["hget"] = s.Hget
    s.cmdMap["hdel"] = s.Hdel
    s.cmdMap["hlen"] = s.Hlen

//for list
    s.cmdMap["lpush"] = s.Lpush
    s.cmdMap["lrange"] = s.Lrange

//for set

//for zset

//for db
    s.cmdMap["select"] = s.Select
    s.cmdMap["keys"] = s.Keys
    s.cmdMap["del"] = s.Del
}
