package server

import (
	"github.com/fvbock/endless"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"os"
	"qiniupkg.com/x/errors.v7"
	"bigbigTravel/component/http/http_middleware"
	"bigbigTravel/component/logger"
)

type HttpServer struct {
	Listen 	string
	Router  *gin.Engine
}

func NewServer(listen string) *HttpServer {
	server :=  &HttpServer{
		Listen: listen,
		Router: gin.New(),
	}
	server.init()
	return server
}

func (s *HttpServer) init() error {
	if s == nil {
		return errors.New("httpServer is nil")
	}

	router := s.Router
	corsConf := cors.DefaultConfig()
	corsConf.AllowCredentials = true
	corsConf.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "authorization"}
	corsConf.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	corsConf.AllowAllOrigins = true

	router.Use(cors.New(corsConf), http_middleware.AccessLogger(), http_middleware.ResponseLogger(), http_middleware.RecoveryLogger(), http_middleware.CheckParams)
	for method, paths := range http_middleware.HttpActionMap {
		for path, handler := range paths {
			switch method {
			case http_middleware.MethodGet:
				router.GET(path, handler)
			case http_middleware.MethodPost:
				router.POST(path, handler)
			case http_middleware.MethodAll:
				router.GET(path, handler)
				router.POST(path, handler)
			default:
				logger.System("server", "server run failed, invalid routing")
				os.Exit(1)
			}
		}
	}

	return nil
}



func (s *HttpServer) Run() error {
	ss := endless.NewServer(s.Listen, s.Router)
	err := ss.ListenAndServe()
	if err != nil {
		logger.System("server", "server listen error:", err.Error())
		os.Exit(1)
	}
	return nil
}