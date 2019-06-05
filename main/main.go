package main

import (
	"bigbigTravel/common/records"
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
	"strings"
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
	updateResourceQiniuHost()
	server := server.NewServer(conf.Config.Server.ServerPort)
	server.Run()
}

func updateResourceQiniuHost() {
	//刷七牛云url host
	db := mysql.GetInstance(false)
	resourceRecordList := db.Find(records.RecordNameResource).Select("*").
		Where("qiniu_url", "like", "http://prfcg2v7u.bkt.clouddn.com%").Execute().FetchAll()
	if resourceRecordList != nil && resourceRecordList.Len() > 0 {
		for _, resourceRecord := range resourceRecordList.AllRecord() {
			resource := resourceRecord.(*records.Resource)
			resource.QiniuUrl = strings.Replace(resource.QiniuUrl, "http://prfcg2v7u.bkt.clouddn.com", "http://qiniu.ruan89.cn", -1)
			db.SaveRecord(resource)
		}
	}
}