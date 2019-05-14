package exception


var (
	ExceptionSeverPanic    	  = DefineException(500, "程序崩溃", "server panic")

	ExceptionInvalidParams    = DefineException(1001, "非法参数", "invalid parameter")
	ExceptionTokenError       = DefineException(1002, "token错误", "token error")
	ExceptionTokenRedisNotExist  = DefineException(1003, "token redis cache不存在", "token redis cache not exist")
	ExceptionTokenNotMatch		= DefineException(1004, "token与redis记录不一致", "token not match")
	ExceptionWxCodeError		= DefineException(1005, "wxCode错误", "wxCode error")
	ExceptionVerifyCodeError    = DefineException(1006, "短信验证码错误", "sms code error")
	ExceptionWxCodeParseError	= DefineException(1007, "微信code解析错误", "wx code error")
	ExceptionWxEncryptedDataParseError	= DefineException(1008, "微信encrypted data解析错误", "wx encrypted data parse error")
	ExceptionDBError	= DefineException(1009, "db操作异常", "db op error")
	ExceptionWxUnifiedOrderFailed	= DefineException(1010, "微信统一下单失败", "wx unified order failed")
	ExceptionWxPayNotifyXmlParseError	= DefineException(1011, "微信支付xml解析失败", "wx xml parse failed")


	ExceptionMissAdmin	= DefineException(2001, "找不到该管理员", "miss admin")
	ExceptionPasswordMismatch	= DefineException(2002, "密码不匹配", "admin password mismatch")
	ExceptionUnexpectedOrderType	= DefineException(2003, "找不到订单类型", "unexpected order type")
	ExceptionResourceUploadError	= DefineException(2004, "资源上传错误", "resource upload error")

)

type CommonError struct {
	displayMsg string
	logMsg     string
	detailMsg  string
	code       int
}

func (e *CommonError) Code() int {
	return e.code
}

func (e *CommonError) DisplayMsg() string {
	return e.displayMsg
}

func (e *CommonError) LogMsg() string {
	return e.logMsg
}

func (e *CommonError) Error() string {
	return e.detailMsg
}
