package methods

import (
	"bigbigTravel/component/wxpay2"
	"bigbigTravel/conf"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/liyoung1992/wechatpay"
	"net/http"
	"strings"
)

func init() {

}

func ParseWxCode(code string, wx *conf.WxConfig) (map[string]string, error) {
	if wx == nil || wx.AppId == "" || wx.AppSecretKey == "" || wx.CodeToSessionURL == "" {
		return nil, errors.New(fmt.Sprintf("wx config is empty %v", wx))
	}
	URL := strings.Replace(wx.CodeToSessionURL, "{appid}", wx.AppId, -1)
	URL = strings.Replace(URL, "{appsecret}", wx.AppSecretKey, -1)
	URL = strings.Replace(URL, "{code}", code, -1)

	resp, err := http.Get(URL)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("wx code to session api req fail %v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("wx code to session api resp status not 200: %v", resp.StatusCode))
	}
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("wx code to session resp decode fail %v", err))
	}
	if _, ok := data["session_key"]; !ok {
		return nil, errors.New(fmt.Sprintf("not found session key: %v", data))
	}
	res := make(map[string]string, 0)
	res["openid"] = data["openid"].(string)
	res["session_key"] = data["session_key"].(string)
	return res, nil
}

type WxUserInfo struct {
	OpenId    string `json:"openId" form:"openId"`
	NickName  string `json:"nickName" form:"nickName"`
	Gender    int    `json:"gender" form:"gender"` //0-未知  1-男 2-女
	City      string `json:"city" form:"city"`
	Province  string `json:"province" form:"province"`
	Country   string `json:"country" form:"country"`
	AvatarUrl string `json:"avatarUrl" form:"avatarUrl"`
	UnionId   string `json:"unionId" form:"unionId"`
}

//v2
func ParseWxEncryptedData(encryptedData, sessionKey, iv string) ([]byte, error) {
	if len(sessionKey) != 24 {
		return nil, errors.New("session key length is error")
	}

	aesKey, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("base64 decode session key failed with error:%s", err.Error()))
	}

	if len(iv) != 24 {
		return nil, errors.New("iv length is error")
	}
	aesIv, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("base64 decode iv failed with error:%s", err.Error()))
	}

	aesCipherText, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("base64 decode encrypted data failed with error:%s", err.Error()))
	}
	aesPlantText := make([]byte, len(aesCipherText))

	aesBlock, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("create new cipher failed with error:%s", err.Error()))
	}

	mode := cipher.NewCBCDecrypter(aesBlock, aesIv)
	mode.CryptBlocks(aesPlantText, aesCipherText)
	aesPlantText = PKCS7UnPadding(aesPlantText)
	return aesPlantText, nil
}

// PKCS7UnPadding return unpadding []Byte plantText
func PKCS7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	if length > 0 {
		unPadding := int(plantText[length-1])
		return plantText[:(length - unPadding)]
	}
	return plantText
}

//v1
func aesDecrypt(cipherBytes, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	dst := make([]byte, len(cipherBytes))
	blockModel.CryptBlocks(dst, cipherBytes)
	dst = pkcs7UnPadding(dst, block.BlockSize())
	return dst, nil
}

func pkcs7UnPadding(dst []byte, blockSize int) []byte {
	length := len(dst)
	unpadding := int(dst[length-1])
	return dst[:(length - unpadding)]
}

//
//type UnifiedOrderRequest struct {
//	Appid    		string 		`json:"appid" form:"appid"`
//	Mchid    		string 		`json:"mch_id" form:"mch_id"`
//	//DeviceInfo    string 		`json:"device_info" form:"device_info"`
//	NonceStr    	string 		`json:"nonce_str" form:"nonce_str"`
//	Sign    		string 		`json:"sign" form:"sign"`
//	//SignType    	string 		`json:"sign_type" form:"sign_type"`
//	Body    		string 		`json:"body" form:"body"`
//	//Detail    	string 		`json:"detail" form:"detail"`
//	//Attach    	string 		`json:"attach" form:"attach"`
//	OutTradeNo    	string 		`json:"out_trade_no" form:"out_trade_no"`
//	//FeeType    	string 		`json:"fee_type" form:"fee_type"`
//	TotalFee    	int 		`json:"total_fee" form:"total_fee"`
//	SpbillCreateIp  string 		`json:"spbill_create_ip" form:"spbill_create_ip"`
//	//TimeStart    	string 		`json:"time_start" form:"time_start"`
//	//TimeExpire    string 		`json:"time_expire" form:"time_expire"`
//	//GoodsTag    	string 		`json:"goods_tag" form:"goods_tag"`
//	NotifyUrl    	string 		`json:"notify_url" form:"notify_url"`
//	TradeType    	string 		`json:"trade_type" form:"trade_type"`
//	//ProductId    	string 		`json:"product_id" form:"product_id"`
//	LimitPay    	string 		`json:"limit_pay" form:"limit_pay"`
//	OpenId    		string 		`json:"openid" form:"openid"`
//	//Receipt    	string 		`json:"receipt" form:"receipt"`
//	SceneInfo    	string 		`json:"scene_info" form:"scene_info"`
//}
func UnifiedOrder(wc *conf.WxConfig, outTradeNo string, clientIp string, openid string, price int) (wxpay2.Params, error) {
	//client := wxpay.NewClient(wxpay.NewAccount(wc.AppId, wc.MchId, wc.ApiKey, false).SetCertData(wc.PayCertDataPath))
	account := wxpay2.NewAccount(wc.AppId, wc.MchId, wc.ApiKey, false)
	account.SetCertData(wc.PayCertDataPath)
	client := wxpay2.NewClient(account)

	params := make(wxpay2.Params)
	params.SetString("body", "test").
		SetString("out_trade_no", outTradeNo).
		SetInt64("total_fee", int64(price)).
		SetString("spbill_create_ip", clientIp).
		SetString("notify_url", wc.NotifyUrl).
		SetString("trade_type", "JSAPI").
		SetString("openid", openid)
	resp, err := client.UnifiedOrder(params)
	if err != nil {
		return nil, err
	}

	return resp, nil

	//signStr := client.Sign(resp)
	//return signStr, nil
}
//
//
//// 订单查询
//params := make(wxpay.Params)
//params.SetString("out_trade_no", "3568785")
//p, _ := client.OrderQuery(params)
//
//// 退款
//params := make(wxpay.Params)
//params.SetString("out_trade_no", "3568785").
//SetString("out_refund_no", "19374568").
//SetInt64("total_fee", 1).
//SetInt64("refund_fee", 1)
//p, _ := client.Refund(params)
//
//// 退款查询
//params := make(wxpay.Params)
//params.SetString("out_refund_no", "3568785")
//p, _ := client.RefundQuery(params)
//
//// 创建支付账户
//account := wxpay.NewAccount("appid", "mchid", "apiKey")
//
//// 设置证书
//account.SetCertData("证书地址")
//
//// 新建微信支付客户端
//client := wxpay.NewClient(account, false) // sandbox环境请传true
//
//// 设置http请求超时时间
//client.SetHttpConnectTimeoutMs(2000)
//
//// 设置http读取信息流超时时间
//client.SetHttpReadTimeoutMs(1000)
//
//// 更改签名类型
//client.SetSignType(HMACSHA256)
//
//// 设置支付账户
//client.setAccount(account)
//
//// 签名
//signStr := client.Sign(params)
//
//// 校验签名
//b := client.ValidSign(params)
//
//// 支付或退款返回成功信息
//return wxpay.Notifies{}.OK()
//
//// 支付或退款返回失败信息
//return wxpay.Notifies{}.NotOK("支付失败或退款失败了")


func WxPayNotify(c *gin.Context, body []byte, wc *conf.WxConfig) (string, error) {
	var req wechatpay.PayNotifyResult
	err := xml.Unmarshal(body, &req)
	if err != nil {
		return "", err
	}
	//var reqMap map[string]interface{}
	//reqMap = make(map[string]interface{}, 0)

	//reqMap["return_code"] = req.ReturnCode
	//reqMap["return_msg"] = req.ReturnMsg
	//reqMap["appid"] = req.AppId
	//reqMap["mch_id"] = req.MchId
	//reqMap["nonce_str"] = req.NonceStr
	//reqMap["result_code"] = req.ResultCode
	//reqMap["openid"] = req.OpenId
	//reqMap["is_subscribe"] = req.IsSubscribe
	//reqMap["trade_type"] = req.TradeType
	//reqMap["bank_type"] = req.BankType
	//reqMap["total_fee"] = req.TotalFee
	//reqMap["fee_type"] = req.FeeType
	//reqMap["cash_fee"] = req.CashFee
	//reqMap["cash_fee_type"] = req.CashFeeType
	//reqMap["transaction_id"] = req.TransactionId
	//reqMap["out_trade_no"] = req.OutTradeNo
	//reqMap["attach"] = req.Attach
	//reqMap["time_end"] = req.TimeEnd

	return req.OutTradeNo, nil
}