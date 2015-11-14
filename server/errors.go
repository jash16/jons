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
    oneKey string = ":1\r\n"
    twoKey string = ":2\r\n"
    threeKey string = ":3\r\n"
    fourKey string = ":4\r\n"
    fiveKey string = ":5\r\n"
    sixKey string = ":6\r\n"
    sevenKey string = ":7\r\n"
    eightKey string = ":8\r\n"
    zeroLine string = "*0\r\n"
    oneLine string="*1\r\n"
    twoLine string = "*2\r\n"
)
