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
	"bigbigTravel/component/wxpay"
	"bigbigTravel/conf"
	"bigbigTravel/consts"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strconv"
	"strings"
	"errors"
	"time"
)

func init() {
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/login", customerLogin)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/register", customerRegister)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/sendSmsCode", customerSendSmsCode)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getMainTagList", customerGetMainTagList)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getProductsByMainTag", customerGetProductsByMainTag)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getProductDetail", customerGetProductDetail)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/privateOrder", customerPrivateOrder)

	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/wxPayDeposit", customerWxPayDeposit)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/wxPayNotify", customerWxPayNotify)

	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/needAuthorize", customerNeedAuthorize)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/authorize/name", customerAuthorizeName)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/authorize/mobile", customerAuthorizeMobile)

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
	WxCode          string `json:"wxCode" form:"wxCode"`
}
func customerRegister(c *gin.Context) {
	req := new(CustomerRegisterRequest)
	httplib.Load(c, req)
	uuid := c.GetString("requestId")

	//if !sms.Match(req.Mobile, req.Code) {
	//	httplib.Failure(c, exception.ExceptionVerifyCodeError)
	//	return
	//}

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
	customerId := db.Insert(records.RecordNameCustomer).Columns("open_id", "abandon").
		Value(openId, 0).Execute().LastInsertId()
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
	productRecordList := db.Find(records.RecordNameProduct).Select("*").
		Where("main_tags", "like", "%"+req.MainTag+"%").
		Where("show", "=", 1).Execute().FetchAll()
	resp := make([]*CustomerGetProductsByMainTagResponseItem, 0)
	if productRecordList != nil && productRecordList.Len() > 0 {
		for _, productRecord := range productRecordList.AllRecord() {
			product := productRecord.(*records.Product)
			item := new(CustomerGetProductsByMainTagResponseItem)
			//取第一个title resource id， 获取图片资源url
			titleResourceList := strings.Split(product.TitleResourceIds, ",")
			if len(titleResourceList) > 0 {
				firstResourceId, err := strconv.Atoi(titleResourceList[0])
				if err == nil {
					if resourceRecord := db.FindOneByPrimary(records.RecordNameResource, firstResourceId); resourceRecord != nil {
						item.ImageUrl = resourceRecord.(*records.Resource).QiniuUrl
					}
				}
			}
			item.ProductId = product.ProductId
			item.ProductName = product.ProductName
			item.SubTags = strings.Split(product.SubTags, ",")
			resp = append(resp, item)
		}
	}

	httplib.Success(c, map[string]interface{}{"list":resp})
	return
}

type CustomerGetProductDetailRequest struct {
	ProductId		int			`json:"productId" form:"productId"`
}
type CustomerGetProductDetailResponse struct {
	ProductName			string	`json:"productName" form:"productName"`
	Type				int		`json:"type" form:"type"`
	Destination			string	`json:"destination" form:"destination"`
	Price				int		`json:"price" form:"price"`
	TitleImageUrls		[]string	`json:"titleImageUrls" form:"titleImageUrls"`
	DetailImageUrls		[]string	`json:"detailImageUrls" form:"detailImageUrls"`
	SubTags				[]string	`json:"subTags" form:"subTags"`
}
func customerGetProductDetail(c *gin.Context) {
	req := new(CustomerGetProductDetailRequest)
	resp := new(CustomerGetProductDetailResponse)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	productRecord := db.FindOneByPrimary(records.RecordNameProduct, req.ProductId)
	if productRecord == nil {
		httplib.Success(c, map[string]interface{}{"detail":resp})
		return
	}
	product := productRecord.(*records.Product)
	resp.ProductName = product.ProductName
	resp.Type = product.Type
	resp.Price = product.Price
	resp.Destination = product.Destination
	resp.SubTags = strings.Split(product.SubTags, ",")

	resp.TitleImageUrls = make([]string, 0)
	titleImageIds := strings.Split(product.TitleResourceIds, ",")
	for _, resourceIdStr := range titleImageIds {
		resourceId, _ := strconv.Atoi(resourceIdStr)
		resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
		if resourceRecord != nil {
			resp.TitleImageUrls = append(resp.TitleImageUrls, resourceRecord.(*records.Resource).QiniuUrl)
		}
	}

	resp.DetailImageUrls = make([]string, 0)
	detailImageIds := strings.Split(product.DetailResourceIds, ",")
	for _, resourceIdStr := range detailImageIds {
		resourceId, _ := strconv.Atoi(resourceIdStr)
		resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
		if resourceRecord != nil {
			resp.DetailImageUrls = append(resp.DetailImageUrls, resourceRecord.(*records.Resource).QiniuUrl)
		}
	}

	httplib.Success(c, map[string]interface{}{"detail":resp})
	return
}


type CustomerPrivateOrderRequest struct {
	Destination         string 		`json:"destination" form:"destination"`
	Mobile          	string 		`json:"mobile" form:"mobile"`
	Name          		string 		`json:"name" form:"name"`
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
	db.Insert(records.RecordNamePrivateOrder).Columns("customer_id", "mobile", "name", "destination", "withdraw", "handled").
		Value(customerId, req.Mobile, req.Name, req.Destination, 0, 0).Execute()
	httplib.Success(c)
	return
}

type CustomerPayDepositRequest struct {
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
	db := mysql.GetInstance(false)

	customer := db.FindOneByPrimary(records.RecordNameCustomer, customerId).(*records.Customer)

	orderId := db.Insert(records.RecordNameNormalOrder).Columns("customer_id", "mobile", "name", "product_id", "payed", "withdraw").
		Value(customerId, customer.Mobile, customer.CustomerName, req.ProductId, 0, 0).Execute().LastInsertId()
	params, err := methods.UnifiedOrder(conf.Config.Wx, gen32TradeNo(strconv.Itoa(orderId)), c.ClientIP(), customer.OpenId)
	if err != nil {
		httplib.Failure(c, exception.ExceptionWxUnifiedOrderFailed, err.Error())
		return
	}

	resp := map[string]string{}
	resp["appId"] = params["appid"]
	resp["timeStamp"] = strconv.Itoa(int(time.Now().Unix()))
	resp["nonceStr"] = wxpay.NonceStr()
	resp["package"] = fmt.Sprintf("prepay_id=%s", params["prepay_id"])
	resp["signType"] = "MD5"
	resp["paySign"] = wxpay.Resign(resp, conf.Config.Wx.ApiKey)

	httplib.Success(c, resp)
	return
}


func customerWxPayNotify(c *gin.Context) {
	uuid := c.GetString("requestId")
	body, _ := ioutil.ReadAll(c.Request.Body)

	fmt.Println("wx notify")

	outTradeNo, err := methods.WxPayNotify(c, body, conf.Config.Wx)
	if err != nil {
		logger.Error("customerWxPayNotify", uuid, err.Error())
		return
	}
	fmt.Println("notify outTrade no:", outTradeNo)
	db := mysql.GetInstance(false)
	orderId, err := strconv.Atoi(parse32TradeNo(outTradeNo))
	if err != nil {
		logger.Error("customerWxPayNotify", uuid, err.Error())
		return
	}
	fmt.Println("notify normal order id:", orderId)
	normalOrderRecord :=db.FindOneByPrimary(records.RecordNameNormalOrder, orderId)
	if normalOrderRecord == nil {
		logger.Error("customerWxPayNotify", uuid, errors.New("miss normal order record"))
			return
	}
	normalOrder := normalOrderRecord.(*records.NormalOrder)
	if normalOrder.Payed == 0 {
		normalOrder.Payed = 1
		db.SaveRecord(normalOrder)
	}
	httplib.Success(c)
	return
}


func gen32TradeNo(origin string) string{
	l := len([]rune(origin))
	zeroCnt := 32 - l - 1
	result := "1"
	for i := 0; i < zeroCnt; i++ {
		result += "0"
	}
	result += origin
	return result
}

func parse32TradeNo(origin string) string{
	result := strings.TrimPrefix(origin, "1")
	return strings.TrimLeft(result, "0")
}


func customerNeedAuthorize(c *gin.Context) {
	var customerId int
	var ok bool
	if customerId, ok = methods.ParseHttpContextToken(c, consts.Customer); !ok {
		return
	}
	db := mysql.GetInstance(false)
	customerRecord := db.FindOneByPrimary(records.RecordNameCustomer, customerId)
	if customerRecord == nil {
		httplib.Success(c, map[string]interface{}{"need":1})
	} else {
		customer := customerRecord.(*records.Customer)
		if customer.Mobile == "" || customer.CustomerName == "" {
			httplib.Success(c, map[string]interface{}{"need":1})
		} else {
			httplib.Success(c, map[string]interface{}{"need":0})
		}
	}

	return
}

type CustomerAuthorizeNameRequest struct {
	Name			string `json:"name" form:"name"`
	WxCode			string 	`json:"wxCode" form:"wxCode"`
	EncryptedData   string `json:"encryptedData" form:"encryptedData"`
	EncryptedDataIv string `json:"encryptedDataIv" form:"encryptedDataIv"`
}
type WxPhoneNumberInfo struct {
	PhoneNumber			string 	`json:"phoneNumber" form:"phoneNumber"`
	PurePhoneNumber		string 	`json:"purePhoneNumber" form:"purePhoneNumber"`
}
func customerAuthorizeName(c *gin.Context) {
	var customerId int
	var ok bool
	if customerId, ok = methods.ParseHttpContextToken(c, consts.Customer); !ok {
		return
	}
	req := new(CustomerAuthorizeNameRequest)
	httplib.Load(c, req)

	db := mysql.GetInstance(false)
	db.Update(records.RecordNameCustomer).Set("name", req.Name).
		Where("customer_id", "=", customerId).Execute()
	httplib.Success(c)
	return
}

type CustomerAuthorizeMobileRequest struct {
	WxCode			string 	`json:"wxCode" form:"wxCode"`
	EncryptedData   string `json:"encryptedData" form:"encryptedData"`
	EncryptedDataIv string `json:"encryptedDataIv" form:"encryptedDataIv"`
}
func customerAuthorizeMobile(c *gin.Context) {
	var customerId int
	var ok bool
	if customerId, ok = methods.ParseHttpContextToken(c, consts.Customer); !ok {
		return
	}
	req := new(CustomerAuthorizeMobileRequest)
	httplib.Load(c, req)
	wxMap, err := methods.ParseWxCode(req.WxCode, conf.Config.Wx)
	if err != nil {
		httplib.Failure(c, exception.ExceptionWxCodeParseError)
		return
	}
	sessionKey := wxMap["session_key"]
	bytes, err := methods.ParseWxEncryptedData(req.EncryptedData, sessionKey, req.EncryptedDataIv)
	if err != nil {
		httplib.Failure(c, exception.ExceptionWxEncryptedDataParseError)
	}
	phoneInfo := new(WxPhoneNumberInfo)
	json.Unmarshal(bytes, phoneInfo)

	db := mysql.GetInstance(false)
	db.Update(records.RecordNameCustomer).Set("mobile", phoneInfo.PurePhoneNumber).
		Where("customer_id", "=", customerId).Execute()
	httplib.Success(c)
	return
}
