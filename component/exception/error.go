package exception

import "net/http"

var (
	ErrorGoroutinePanic		= &RunError{http.StatusInternalServerError,"程序内部错误","server internel error"}
	ErrorMysqlPanic			= &RunError{http.StatusInternalServerError,"程序内部错误","mysql err"}
	ErrorRedisPanic			= &RunError{http.StatusInternalServerError,"程序内部错误","redis err"}
	ErrorKaerPanic          = &RunError{http.StatusInternalServerError,"程序内部错误","kaer err"}
)