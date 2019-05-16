package orm_gen

import (
	"bigbigTravel/component/mysql"
	"testing"
)

func initMysql() {
	mysql.InitConfig(&mysql.MysqlConfig{
		Host:            "127.0.0.1",
		Port:            3389,
		UserName:        "panruibajiu",
		PassWord:        "PanruiBajiu123!@#",
		DBName:          "bigbigtravel",
		MaxConnLifeTime: 3600,
		MaxIdleConns:    32,
		MaxOpenConns:    200,
	})
}

var path = "/Users/ruipan/myProjects/bigbigTravel/common/records"

//func TestGenerateGlobalRecords(t *testing.T) {
//	initMysql()
//	GenerateRecords(path, GenerateGlobal)
//}

func TestGenerateSingleRecords(t *testing.T) {
	initMysql()
	records := []string{"product", "resource"}
	GenerateRecords(path, GenerateSingle, records...)
}

