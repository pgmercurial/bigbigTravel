package http_middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"bigbigTravel/component/logger"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ResponseLogger() gin.HandlerFunc {
	return  func(context * gin.Context){
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: context.Writer}
		context.Writer = blw
		context.Next()
		respString := blw.body.String()
		if  respString == ""{
			return
		}
		logger.Info("http_response", context.GetString("requestId"), respString)
	}
}

