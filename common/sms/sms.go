package sms

import (
	"bigbigTravel/component/redis"
	"bigbigTravel/consts"
	"fmt"
	"math/rand"
	"time"
)

func init() {

}

func Match(mobile string, verifyCode string) bool { 	//校验短信校验码
	rds := redis.GetInstance()
	defer rds.Close()
	smsKey := cacheKey(mobile)
	cacheCode := rds.Get(smsKey)
	if verifyCode == cacheCode {
		return true
	} else {
		return false
	}
}

func cacheKey(mobile string) string {
	str := mobile
	return redis.GenKey(consts.AppName, true, "string", "user", "smsCode", str)
}

// 生成四位随机验证码
func getRandomCode() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%04v", rnd.Int31n(10000))
	return vcode
}

// 发送普通验证码短信  //todo
func SendVerifyCode (mobile string) {
	code := getRandomCode()
	rds := redis.GetInstance()
	defer rds.Close()
	cacheKey := cacheKey(mobile)
	rds.Set(cacheKey, code, 15*60)

	//go sendMessage(content, mobile, "/sms/send-sms", 200, false)
}


