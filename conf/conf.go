package conf

import(
	"bigbigTravel/component/mysql"
	"bigbigTravel/component/redis"
	"github.com/BurntSushi/toml"
	"bigbigTravel/component/logger"
	"bigbigTravel/component/qiniu"
)

func init() {

}

type AppConfig struct {
	Server      *ServerConfig
	Logger      *logger.LogConfig
	Mysql       *mysql.MysqlConfig
	Redis       *redis.RedisConfig
	Qiniu       *qiniu.QiniuConfig
	//Params                *Params
	Wx          *WxConfig

}

type ServerConfig struct {
	ServerIP     string
	ServerPort   string
	PidFile      string
	Env 		 string
}

type WxConfig struct {
	AppId             string
	AppSecretKey      string
	CodeToSessionURL  string
	TemplateURL       string
	AppAccessTokenURL string

	MchId 				string
	UnifiedOrderURL		string
	NotifyUrl			string
}

var Config AppConfig


func LoadConfig(tomlpath string) bool {
	if _, err := toml.DecodeFile(tomlpath, &Config); err != nil {
		return false
	}
	return true
}