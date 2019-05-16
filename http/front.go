package http

import (
	"bigbigTravel/common/methods"
	"bigbigTravel/common/records"
	"bigbigTravel/common/sms"
	"bigbigTravel/component/exception"
	"bigbigTravel/component/http/http_middleware"
	"bigbigTravel/component/http/httplib"
	"bigbigTravel/component/logger"
	"bigbigTravel/component/mysql"
	"bigbigTravel/conf"
	"bigbigTravel/consts"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strconv"
	"strings"
	"errors"
)

func init() {
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/login", customerLogin)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/register", customerRegister)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/sendSmsCode", customerSendSmsCode)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getMainTagList", customerGetMainTagList)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getProductsByMainTag", customerGetProductsByMainTag)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/privateOrder", customerPrivateOrder)

	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/wxPayDeposit", customerWxPayDeposit)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/wxPayNotify", customerWxPayNotify)

}

type CustomerLoginRequest struct {
	WxCode      string `json:"wxCode" form:"wxCode"`
}
func customerLogin(c *gin.Context) {
	req := new(CustomerLoginRequest)
	httplib.Load(c, req)
	if req.WxCode == "" {
		httplib.Failure(c, exception.ExceptionTokenError)
		return
	}

	wxMap, err := methods.ParseWxCode(req.WxCode, conf.Config.Wx)
	if err != nil {
		httplib.Failure(c, exception.ExceptionTokenError)
		return
	}
	openId := wxMap["openid"]

	db := mysql.GetInstance(false)
	customerRecord := db.Find(records.RecordNameCustomer).Select("*").Where("open_id", "=", openId).Execute().Fetch()
	var customerId int
	if customerRecord != nil {
		customerId = customerRecord.(*records.Customer).CustomerId
		token, _ := methods.GenUserToken(customerId, consts.Customer)
		httplib.Success(c, map[string]interface{}{"token":token, "found":1})
		return
	} else {
		httplib.Success(c, map[string]interface{}{"token":"", "found":0})
		return
	}
}

type CustomerRegisterRequest struct {
	Mobile          string `json:"mobile" form:"mobile"`
	Code            string `json:"code" form:"code"`
	WxCode          string `json:"wxCode" form:"wxCode"`
	//EncryptedData   string `json:"encryptedData" form:"encryptedData"`
	//EncryptedDataIv string `json:"encryptedDataIv" form:"encryptedDataIv"`
}
func customerRegister(c *gin.Context) {
	req := new(CustomerRegisterRequest)
	httplib.Load(c, req)
	uuid := c.GetString("requestId")

	if !sms.Match(req.Mobile, req.Code) {
		httplib.Failure(c, exception.ExceptionVerifyCodeError)
		return
	}

	wxMap, parseErr := methods.ParseWxCode(req.WxCode, conf.Config.Wx)
	if parseErr != nil {
		logger.Error("customerRegister", uuid, fmt.Sprintf("wxCode parse error:%s", parseErr.Error()))
		httplib.Failure(c, exception.ExceptionWxCodeParseError)
		return
	}
	openId := wxMap["openid"]
	//sessionKey := wxMap["session_key"]
	//wxUserInfo, err := methods.ParseWxEncryptedData(req.EncryptedData, sessionKey, req.EncryptedDataIv)
	//if err != nil {
	//	httplib.Failure(c, exception.ExceptionWxEncryptedDataParseError)
	//}
	db := mysql.GetInstance(false)
	//customerId := db.Insert(records.RecordNameCustomer).Columns("open_id", "name", "head_photo", "mobile", "abandon").
	//	Value(openId, wxUserInfo.NickName, wxUserInfo.AvatarUrl, req.Mobile, 0).Execute().LastInsertId()
	customerId := db.Insert(records.RecordNameCustomer).Columns("open_id", "mobile", "abandon").
		Value(openId, req.Mobile, 0).Execute().LastInsertId()
	if customerId <= 0 {
		httplib.Failure(c, exception.ExceptionDBError)
	}
	token, _ := methods.GenUserToken(customerId, consts.Customer)
	httplib.Success(c, map[string]interface{}{"token":token})
	return
}

type CustomerSendSmsCodeRequest struct {
	Mobile          string `json:"mobile" form:"mobile"`
}
func customerSendSmsCode(c *gin.Context) {
	req := new(CustomerSendSmsCodeRequest)
	httplib.Load(c, req)
	sms.SendVerifyCode(req.Mobile)
	httplib.Success(c)
	return
}

func customerGetMainTagList(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Customer); !success {
		return
	}
	db := mysql.GetInstance(false)
	sysconfRecord:= db.Find(records.RecordNameSysConf).Select("*").Where("enable", "=", 1).Execute().Fetch()
	if sysconfRecord == nil {
		httplib.Success(c, map[string]interface{}{"list":[]string{}})
		return
	} else {
		sysConf := sysconfRecord.(*records.SysConf)
		mainTagsStr := sysConf.MainTags
		mainTagList := strings.Split(mainTagsStr, ",")
		httplib.Success(c, map[string]interface{}{"list":mainTagList})
		return
	}
}

type CustomerGetProductsByMainTagRequest struct {
	MainTag          string `json:"mainTag" form:"mainTag"`
}
type CustomerGetProductsByMainTagResponseItem struct {
	ProductId         int 		`json:"productId" form:"productId"`
	ProductName       string 	`json:"productName" form:"productName"`
	ImageUrl          string 	`json:"imageUrl" form:"imageUrl"`
	SubTags           []string 	`json:"subTags" form:"subTags"`
}

func customerGetProductsByMainTag(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Customer); !success {
		return
	}
	req := new(CustomerGetProductsByMainTagRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	//todo:更多搜索条件？
	productRecordList := db.Find(records.RecordNameProduct).Select("*").
		Where("main_tags", "like", "%"+req.MainTag+"%").
		Where("show", "=", 1).Execute().FetchAll()
	resp := make([]*CustomerGetProductsByMainTagResponseItem, 0)
	if productRecordList != nil && productRecordList.Len() > 0 {
		for _, productRecord := range productRecordList.AllRecord() {
			product := productRecord.(*records.Product)
			item := new(CustomerGetProductsByMainTagResponseItem)
			//item.ImageUrl = product.DetailImageUrl //todo:?
			item.ProductId = product.ProductId
			item.ProductName = product.ProductName
			item.SubTags = strings.Split(product.SubTags, ",")
			resp = append(resp, item)
		}
	}

	httplib.Success(c, map[string]interface{}{"list":resp})
	return
}


type CustomerPrivateOrderRequest struct {
	Destination          string `json:"destination" form:"destination"`
}
func customerPrivateOrder(c *gin.Context) {
	var customerId int
	var success bool
	if customerId, success = methods.ParseHttpContextToken(c, consts.Customer); !success {
		return
	}
	req := new(CustomerPrivateOrderRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	db.Insert(records.RecordNamePrivateOrder).Columns("customer_id", "destination", "withdraw", "handled").
		Value(customerId, req.Destination, 0, 0).Execute()
	httplib.Success(c)
	return
}

type CustomerPayDepositRequest struct {
	WxCode		string 	`json:"wxCode" form:"wxCode"`
	ProductId	int 	`json:"productId" form:"productId"`
	ClientIp	string 	`json:"clientIp" form:"clientIp"`

}
func customerWxPayDeposit(c *gin.Context) {  //微信支付定金
	var customerId int
	var success bool
	if customerId, success = methods.ParseHttpContextToken(c, consts.Customer); !success {
		return
	}
	req := new(CustomerPayDepositRequest)
	httplib.Load(c, req)

	_, err := methods.ParseWxCode(req.WxCode, conf.Config.Wx)
	if err != nil {
		httplib.Failure(c, exception.ExceptionWxCodeParseError)
		return
	}

	db := mysql.GetInstance(false)
	orderId := db.Insert(records.RecordNameNormalOrder).Columns("customer_id", "product_id", "valid", "withdraw").
		Value(customerId, req.ProductId, 0, 0).Execute().LastInsertId()
	params, err := methods.UnifiedOrder(conf.Config.Wx, strconv.Itoa(orderId), c.ClientIP())
	if err != nil {
		httplib.Failure(c, exception.ExceptionWxUnifiedOrderFailed, err.Error())
		return
	}

	httplib.Success(c, map[string]string(params))
	return
}


func customerWxPayNotify(c *gin.Context) {
	uuid := c.GetString("requestId")
	body, _ := ioutil.ReadAll(c.Request.Body)

	outTradeNo, err := methods.WxPayNotify(c, body, conf.Config.Wx)
	if err != nil {
		logger.Error("customerWxPayNotify", uuid, err.Error())
		return
	}
	db := mysql.GetInstance(false)
	orderId, err := strconv.Atoi(outTradeNo)
	if err != nil {
		logger.Error("customerWxPayNotify", uuid, err.Error())
		return
	}
	normalOrderRecord :=db.FindOneByPrimary(records.RecordNameNormalOrder, orderId)
	if normalOrderRecord == nil {
		logger.Error("customerWxPayNotify", uuid, errors.New("miss normal order record"))
			return
	}
	normalOrder := normalOrderRecord.(*records.NormalOrder)
	normalOrder.Valid = 1
	db.SaveRecord(normalOrder)
	return
}