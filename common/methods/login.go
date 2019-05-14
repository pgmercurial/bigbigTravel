package methods

import(
	"bigbigTravel/component/helper"
	"bigbigTravel/consts"
	"errors"
)

func init() {

}

func VerifyPassword(origin, encrypted string) error {
	if encrypted != Md5Password(origin) {
		return errors.New("密码不正确")
	}
	return nil
}

//生成密码
func Md5Password(origin string) string {
	return helper.Md5(origin + consts.PasswordSalt)
}