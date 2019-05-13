package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"bigbigTravel/component/http/httplib"
	"bigbigTravel/component/http/http_middleware"
)

func init() {
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "test", test)
}

func test(c *gin.Context) {
	fmt.Println("http test called!")
	httplib.Success(c)
}

