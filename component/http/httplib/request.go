package httplib

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
)

func Load(context *gin.Context, obj interface{}) {
	method := context.Request.Method
	var err error
	if method == "GET" {
		err = loadDefault(context, obj)
	} else {
		switch context.Request.Header.Get("content-type") {
		case "application/json;charset=UTF-8":
			fallthrough
		case "application/json":
			err = loadJson(context, obj)
			break
		default: //case MIMEPOSTForm, MIMEMultipartPOSTForm:
			err = loadDefault(context, obj)
		}
	}
	if err != nil {
		panic(errors.New("httplib request load failed"))
	}
}

func loadDefault(context *gin.Context, obj interface{}) error {
	err := context.Bind(obj)
	formString, _ := json.Marshal(obj)
	SetRequestBodyString(context, formString)
	if err != nil {
		return err
	}
	return nil
}

func loadJson(context *gin.Context, obj interface{}) error {
	body, ok := context.Get("requestBody")
	if !ok {
		return errors.New("json request failed, empty body")
	}
	if err := json.Unmarshal(body.([]byte), obj); err != nil {
		return err
	}
	SetRequestBodyString(context, body.([]byte))
	return nil
}

func SetRequestBodyString(context *gin.Context, data []byte) {
	context.Set("requestBody", string(data))
}

func GetRequestBodyString(context *gin.Context) string {
	return context.GetString("requestBody")
}
