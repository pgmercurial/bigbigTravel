package httplib

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
	"bigbigTravel/component/exception"
)

//===========Response=============
type ApiResponse struct {
	Code  		int			`json:"code"`
	Msg   		string		`json:"msg"`
	RequestId	string		`json:"requestId"`
	Data  		interface{}	`json:"data"`
}

func Success(context *gin.Context, data ...interface{}){
	response := ApiResponse{}
	response.Code = 0
	response.Msg = "success"
	response.RequestId = context.GetString("requestId")

	if len(data) == 0 {
		response.Data = struct {
		}{}
	}else {
		response.Data = data[0]
	}
	context.Set("responseCode", 0)
	context.JSON(http.StatusOK, response)
}

func Failure(context *gin.Context, responseError *exception.RunException, detailMsg ...string)  {
	response := ApiResponse{}
	response.Code = responseError.Code
	response.Msg = responseError.DisplayMsg
	response.RequestId = context.GetString("requestId")
	response.Data = map[string]interface{}{
		"detail err msg": detailMsg,
	}
	context.Set("responseCode", responseError.Code)
	context.Set("exceptionMsg", fmt.Sprintf("%s:\"%s\"", responseError.LogMsg, detailMsg))
	context.JSON(http.StatusOK, response)
}

func Error(context *gin.Context, responseError *exception.RunException, errMsg ...string)  {
	response := ApiResponse{}
	response.Code = responseError.Code
	response.Msg = responseError.DisplayMsg
	response.RequestId = context.GetString("requestId")
	response.Data = struct {
	}{}
	context.Set("responseCode", responseError.Code)
	context.Set("errorMsg", fmt.Sprintf("%s:\"%s\"", responseError.LogMsg, errMsg))
	context.JSON(http.StatusOK, response)
}
