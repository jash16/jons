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
    s.cmdMap["hgetall"] = s.Hgetall
    s.cmdMap["hkeys"] = s.Hkeys
    s.cmdMap["hexists"] = s.Hexists
    s.cmdMap["hvals"] = s.Hvals
    s.cmdMap["hmget"] = s.Hmget
    s.cmdMap["hmset"] = s.Hmset

//for list
    s.cmdMap["lpush"] = s.Lpush
    s.cmdMap["lrange"] = s.Lrange

//for set

//for zset

//for db
    s.cmdMap["select"] = s.Select
    s.cmdMap["keys"] = s.Keys
    s.cmdMap["del"] = s.Del

    s.cmdMap["expire"] = s.Expire
    s.cmdMap["pexpire"] = s.Pexpire

//for sub pub
    s.cmdMap["subscribe"] = s.Subscribe
    s.cmdMap["publish"] = s.Publish

//for persist
    s.cmdMap["bgsave"] = s.Bgsave
}
