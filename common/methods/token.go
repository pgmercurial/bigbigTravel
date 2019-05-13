package methods

import(
	"bigbigTravel/component/exception"
	"bigbigTravel/component/helper"
	"bigbigTravel/component/http/httplib"
	"bigbigTravel/component/mysql"
	"bigbigTravel/component/redis"
	"bigbigTravel/consts"
	"github.com/gin-gonic/gin"
	"math/rand"
	"strconv"
	"strings"
	"errors"
	"bigbigTravel/component/aes"
)

func init() {

}

func ParseHttpContextToken(context *gin.Context) (userId int, success bool) {
	token, ok := context.GetQuery("token")
	if !ok || token == "" {
		token, _ = context.GetPostForm("token")
		if token == "" {
			httplib.Failure(context, exception.ExceptionTokenError)
			return 0, false
		}
	}
	if strings.Contains(token, " ") {
		httplib.Failure(context, exception.ExceptionTokenError)
		return 0, false
	}
	userId, err := getUserIdByToken(token)
	if err != nil {
		if strings.Contains(err.Error(), "token redis empty") {
			httplib.Failure(context, exception.ExceptionTokenRedisNotExist)
		} else if strings.Contains(err.Error(), "token not match") {
			httplib.Failure(context, exception.ExceptionTokenNotMatch)
		} else {
			httplib.Failure(context, exception.ExceptionTokenError)
		}
		return 0, false
	}
	return userId, true
}

func getUserIdByToken(token string) (int, error) {
	if token == "" {
		return 0, errors.New("invalid user, token empty")
	}
	originStr, err := aes.AesDecrypt(token, consts.AES_KEY)
	if nil != err {
		return 0, errors.New("invalid user, decode failed:" + err.Error())
	}

	strArr := strings.Split(originStr, "+")
	if len(strArr) < 2 {

	}
	userId, err := strconv.Atoi(strArr[0])
	if nil != err {
		return 0, errors.New("invalid user, error token:" + err.Error())
	}

	tokenKey := tokenKey(userId)
	rds := redis.GetInstance()
	defer rds.Close()
	tokenValue := rds.Get(tokenKey)
	if tokenValue == "" {
		return 0, errors.New("token redis empty")
	}
	if tokenValue != token {
		return 0, errors.New("invalid user, token not match")
	}

	return userId, nil
}

func GenUserToken(userId int) (string, error) {
	str := strconv.Itoa(userId) + "+" + helper.TimeNow() + "+" + strconv.Itoa(rand.Intn(999))
	token, err := aes.AesEncrypt(str, consts.AES_KEY)
	if err != nil {
		return "", err
	}

	rds := redis.GetInstance()
	defer rds.Close()
	key := tokenKey(userId)
	rds.Set(key, token, 86400)
	return token, nil
}

//func DeleteUserToken(userId int, actor int) {
//	rds := redis.GetInstance()
//	defer rds.Close()
//	key := TokenCacheKey(userId, actor)
//	rds.Del(key)
//}

func tokenKey(userId int) string {
	return redis.GenKey(consts.AppName, true, "string", "user", "usertoken", strconv.Itoa(userId))
}