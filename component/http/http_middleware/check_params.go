package http_middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"bigbigTravel/component/exception"
	"bigbigTravel/component/http/httplib"
)

func CheckParams(context *gin.Context) {
	path := strings.Trim(context.Request.URL.Path, "/")
	checkParams, ok := HttpActionParams[path]
	context.Request.ParseForm()
	context.Request.ParseMultipartForm(32 * 1024 * 1024)
	if ok {
		if context.Request.Method == "POST" && context.Request.Header.Get("content-type") == "application/json" {
			body, ok := context.Get("requestBody")
			if !ok {
				httplib.Failure(context, exception.ExceptionInvalidParams, "json request with empty text post")
				context.Abort()
			}
			strBody := string(body.([]byte))

			mp := make(map[string]interface{})
			json.Unmarshal(body.([]byte), &mp)
			for k, v := range mp {
				if _, ok := context.Request.Form[k]; !ok {
					context.Request.Form[k] = []string{fmt.Sprint(v)}
				} else {
					context.Request.Form[k][0] = fmt.Sprint(v)
				}

			}
			for _, ps := range checkParams {
				exists := false
				for _, p := range ps {
					if strings.Contains(strBody, `"`+p+`"`) {
						exists = true
						break
					}
				}
				if !exists {
					httplib.Failure(context, exception.ExceptionInvalidParams, "need give one or all of params:"+strings.Join(ps, ","))
					context.Abort()
					return
				}
			}
		} else {
			for _, ps := range checkParams {
				exists := false
				for _, p := range ps {
					if _, ok := context.Request.Form[p]; ok {
						exists = true
						break
					}
				}
				if !exists {
					httplib.Failure(context, exception.ExceptionInvalidParams, "need give one or all of params:"+strings.Join(ps, ","))
					context.Abort()
					return
				}
			}
		}
	}
	context.Next()
}
