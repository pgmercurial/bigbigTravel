package http_middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
)

const (
	MethodGet  = "GET"
	MethodPost = "POST"
	MethodAll  = "ALL"
)

var HttpActionParams = make(map[string][][]string)
var HttpActionMap = map[string]map[string]gin.HandlerFunc{}

func registerHttpActionParams(path string, params ...interface{}) {
	path = strings.Trim(path, "/")
	HttpActionParams[strings.Trim(path, "/")] = make([][]string, 0)
	for _, ps := range params {
		v := make([]string, 0)
		if param, ok := ps.(string); ok {
			v = append(v, param)
		} else if params, ok := ps.([]string); ok {
			v = params
		} else {
			panic(errors.New("action params error, path:"+path))
		}
		HttpActionParams[path] = append(HttpActionParams[path], v)
	}
}

func RegisterHttpAction(method, path string, handlerFunc gin.HandlerFunc, params ...interface{}) {
	if _, ok := HttpActionMap[method]; !ok {
		HttpActionMap[method] = make(map[string]gin.HandlerFunc)
	}
	HttpActionMap[method][path] = handlerFunc
	registerHttpActionParams(path, params...)
}
