package common

import (
    "time"
)

func GetMsTime() int64 {
    nowTime := time.Now()
    return nowTime.Unix() * 1000 + int64(nowTime.Nanosecond() / 100000)
}

func GetNsTime() int64 {
    nowTime := time.Now()
    return nowTime.Unix() * 1000000 + int64(nowtTime.Nanosecond())
}
