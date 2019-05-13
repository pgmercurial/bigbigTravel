package main

import (
	"bigbigTravel/component/logger"
	"bigbigTravel/component/mysql"
	"bigbigTravel/component/qiniu"
	"bigbigTravel/component/redis"
	"bigbigTravel/conf"
	_ "bigbigTravel/http"
	"bigbigTravel/server"
	"fmt"
	"os"
	"runtime"
)

func init() {
	if !conf.LoadConfig("conf/conf.toml") {
		os.Exit(1)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	logger.InitConfig(conf.Config.Logger, conf.Config.Server.Env)
	mysql.InitConfig(conf.Config.Mysql)
	redis.InitConfig(conf.Config.Redis)
	qiniu.InitConfig(conf.Config.Qiniu)
}

func main() {
	fmt.Println("server start")
	server := server.NewServer(conf.Config.Server.ServerPort)
	server.Run()
}