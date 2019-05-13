package helper

import (
	"time"
)

const (
	TimeFormatYmdHims 	= "2006-01-02 15:04:05.000"
	TimeFormatYmdHis 	= "2006-01-02 15:04:05"
	TimeFormatYmdHi		= "2006-01-02 15:04"
	TimeFormatYmdH		= "2006-01-02 15"
	TimeFormatYmd		= "2006-01-02"
)

var local,_ = time.LoadLocation("Asia/Shanghai")

var TimeNow = func() string {return time.Now().Format(TimeFormatYmdHis)}

func ParseTime(format string, value string) int {
	v,_ := time.ParseInLocation(format, value, local)
	return int(v.Unix())
}

func FormatTime(format string, unixTime int) string {
	return time.Unix(int64(unixTime), 0).Format(format)
}
