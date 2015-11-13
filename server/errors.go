package server

var (
    ok string = "+OK\r\n"
    wrongType string = "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
    wrongArgs string = "-ERR wrong number of arguments for '%s' command\r\n"
    wrongCommand string = "-ERR unknown command '%s'\r\n"
    wrongDbIdx string = "-ERR invalid DB index\r\n"
    wrongKey string = "$-1\r\n"
    wrongIdx string = "-ERR index out of range\r\n"
    zeroKey string = ":0\r\n"
)
