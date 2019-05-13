package http_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strings"
	"time"
	"bigbigTravel/component/helper"
	"bigbigTravel/component/http/httplib"
	"bigbigTravel/component/logger"
)

func AccessLogger() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Start timer
		requestId := helper.GenerateUUID()
		requestTime := time.Now().Format(helper.TimeFormatYmdHims)
		context.Set("requestId", requestId)
		context.Set("requestTime", requestTime)
		start := time.Now()
		path := context.Request.URL.Path
		raw := context.Request.URL.RawQuery
		// if context.Request.Header.Get("content-type") == "application/json" {
		if strings.Contains(context.Request.Header.Get("content-type"), "application/json") {
			body, _ := ioutil.ReadAll(context.Request.Body)
			logger.Debug("requestBody", requestId, string(body))
			context.Set("requestBody", body)
		}
		context.Next()

		end := time.Now()
		latency := fmt.Sprintf("%v", end.Sub(start))

		clientIP := context.ClientIP()
		method := context.Request.Method
		statusCode := context.Writer.Status()
		comment := context.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}
		postParams := httplib.GetRequestBodyString(context)
		responseCode := context.GetInt("responseCode")

		logger.Info("http_access", requestId, clientIP, method, statusCode, responseCode, latency, path, postParams, comment)

		exceptionMsg := context.GetString("exceptionMsg")
		logger.Warning("exception", requestId, exceptionMsg)
		errorMsg := context.GetString("errorMsg")
		logger.Error("error", requestId, errorMsg)
	}
}
