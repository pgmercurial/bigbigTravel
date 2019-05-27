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
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/login", customerLogin)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/register", customerRegister)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/sendSmsCode", customerSendSmsCode)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getMainTagList", customerGetMainTagList)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getProducts", customerGetProducts)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getProductDetail", customerGetProductDetail)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/privateOrder", customerPrivateOrder)

	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/wxPayDeposit", customerWxPayDeposit)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/wxPayNotify", customerWxPayNotify)

	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/needAuthorize", customerNeedAuthorize)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/authorize/name", customerAuthorizeName)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/authorize/mobile", customerAuthorizeMobile)

	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getProductTitleImages", customerGetProductTitleImages)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getProductsByMainTag", customerGetProductByMainTag)

	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getHeadImages", customerGetHeadImages)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getIntroImages", customerGetIntroImages)

	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "customer/getOrders", customerGetOrders)

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
		fmt.Println(openId,"  ", customerId)
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

type CustomerGetProductsResponseItem1 struct {
	MainTag         string 		`json:"mainTag" form:"mainTag"`
	List			[]*CustomerGetProductsResponseItem2	`json:"list" form:"list"`
}

type CustomerGetProductsResponseItem2 struct {
	ProductId         int 		`json:"productId" form:"productId"`
	Destination       string 	`json:"destination" form:"destination"`
}

func customerGetProducts(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Customer); !success {
		return
	}
	db := mysql.GetInstance(false)
	resp := make([]*CustomerGetProductsResponseItem1, 0)

	sysConfRecord := db.Find(records.RecordNameSysConf).Select("*").Where("enable", "=", 1).Execute().Fetch()
	if sysConfRecord == nil {
		httplib.Success(c, map[string]interface{}{"list":resp})
		return
	}
	sysConf := sysConfRecord.(*records.SysConf)
	mainTags := strings.Split(sysConf.MainTags, ",")

	for _, mainTag := range mainTags {
		if mainTag == "旅行定制" || mainTag == "线路招募" || mainTag == "签证办理" || mainTag == "旅行周边" || mainTag == "自由行" {
			continue
		}
		item1 := new(CustomerGetProductsResponseItem1)
		productRecordList := db.Find(records.RecordNameProduct).Select("*").
			Where("main_tags", "like", "%"+mainTag+"%").
			Where("show", "=", 1).Execute().FetchAll()
		if productRecordList == nil || productRecordList.Len() == 0 {
			continue
		}
		for _, productRecord := range productRecordList.AllRecord() {
			product := productRecord.(*records.Product)
			item2 := new(CustomerGetProductsResponseItem2)
			item2.ProductId = product.ProductId
			item2.Destination = product.Destination
			item1.List = append(item1.List, item2)
		}
		item1.MainTag = mainTag
		resp = append(resp, item1)
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
	product := db.FindOneByPrimary(records.RecordNameProduct, req.ProductId).(*records.Product)

	orderId := db.Insert(records.RecordNameNormalOrder).Columns("customer_id", "mobile", "name", "product_id", "payed", "withdraw").
		Value(customerId, customer.Mobile, customer.CustomerName, req.ProductId, 0, 0).Execute().LastInsertId()
	params, err := methods.UnifiedOrder(conf.Config.Wx, gen32TradeNo(strconv.Itoa(orderId)), c.ClientIP(), customer.OpenId, product.Price)
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

type CustomerGetProductTitleImagesResponseItem struct {
	ProductId			int 	`json:"productId" form:"productId"`
	TitleImage   		string  `json:"titleImage" form:"titleImage"`
}
func customerGetProductTitleImages(c *gin.Context) {
	resp := make([]*CustomerGetProductTitleImagesResponseItem, 0)
	db := mysql.GetInstance(false)
	productIds := db.Find(records.RecordNameProduct).Select("*").Limit(100).Execute().FetchAll().Columns("productId").([]int)
	num := len(productIds)

	fn := func(productId int) {
		product := db.FindOneByPrimary(records.RecordNameProduct, productId).(*records.Product)
		item := new(CustomerGetProductTitleImagesResponseItem)
		item.ProductId = productId
		resources := strings.Split(product.TitleResourceIds, ",")
		if len(resources) == 0 {
			item.TitleImage = ""
		} else {
			resourceId, _ := strconv.Atoi(resources[0])
			resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
			if resourceRecord == nil {
				item.TitleImage = ""
			} else {
				item.TitleImage = resourceRecord.(*records.Resource).QiniuUrl
			}
		}
		resp = append(resp, item)
	}

	if num == 0 {
		httplib.Success(c, map[string]interface{}{"list":resp})
		return
	} else if num <= 3 {
		for _, productId := range productIds {
			fn(productId)
		}
	} else {
		for i := 0; i < 3; i++ {
			ri := rand.Intn(num)
			productId := productIds[ri]
			fn(productId)
			productIds = append(productIds[0:ri], productIds[ri+1:]...)
			num = len(productIds)
		}
	}
	httplib.Success(c, map[string]interface{}{"list":resp})
	return
}

type CustomerGetProductByMainTagRequest struct {
	MainTag		string 	`json:"mainTag" form:"mainTag"`
}
type CustomerGetProductByMainTagResponseItem struct {
	ProductId		int 	`json:"productId" form:"productId"`
	ProductName		string 	`json:"productName" form:"productName"`
	TitleImage		string 	`json:"titleImage" form:"titleImage"`
	Price			int 	`json:"price" form:"price"`

}
func customerGetProductByMainTag(c *gin.Context) {
	req := new(CustomerGetProductByMainTagRequest)
	resp := make([]*CustomerGetProductByMainTagResponseItem, 0)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	productRecordList := db.Find(records.RecordNameProduct).Select("*").
		Where("main_tags", "like", "%"+req.MainTag+"%").
		Where("show", "=", 1).Execute().FetchAll()
	if productRecordList == nil || productRecordList.Len() == 0 {
		httplib.Success(c, map[string]interface{}{"list":resp})
		return
	} else {
		for _, productRecord := range productRecordList.AllRecord() {
			product := productRecord.(*records.Product)
			item := new(CustomerGetProductByMainTagResponseItem)
			item.ProductId = product.ProductId
			item.ProductName = product.ProductName
			item.Price = product.Price
			resources := strings.Split(product.TitleResourceIds, ",")
			if len(resources) == 0 {
				item.TitleImage = ""
			} else {
				resourceId, _ := strconv.Atoi(resources[0])
				resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
				if resourceRecord == nil {
					item.TitleImage = ""
				} else {
					item.TitleImage = resourceRecord.(*records.Resource).QiniuUrl
				}
			}
			resp = append(resp, item)
		}
	}
	httplib.Success(c, map[string]interface{}{"list":resp})
	return
}


func customerGetHeadImages(c *gin.Context) {
	resp := make([]string, 0)
	db := mysql.GetInstance(false)
	sysConfRecord := db.Find(records.RecordNameSysConf).Select("*").Where("enable", "=", 1).Execute().Fetch()
	if sysConfRecord == nil {
		httplib.Success(c, map[string]interface{}{"list":resp})
		return
	}
	images := sysConfRecord.(*records.SysConf).HeadImages
	resourceIds := strings.Split(images, ",")
	for _, resourceIdStr := range resourceIds {
		resourceId, err := strconv.Atoi(resourceIdStr)
		if err != nil {
			continue
		}
		resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
		if resourceRecord == nil {
			continue
		}
		resp = append(resp, resourceRecord.(*records.Resource).QiniuUrl)
	}
	httplib.Success(c, map[string]interface{}{"list":resp})
	return
}

func customerGetIntroImages(c *gin.Context) {
	resp := make([]string, 0)
	db := mysql.GetInstance(false)
	sysConfRecord := db.Find(records.RecordNameSysConf).Select("*").Where("enable", "=", 1).Execute().Fetch()
	if sysConfRecord == nil {
		httplib.Success(c, map[string]interface{}{"list":resp})
		return
	}
	images := sysConfRecord.(*records.SysConf).IntroImages
	resourceIds := strings.Split(images, ",")
	for _, resourceIdStr := range resourceIds {
		resourceId, err := strconv.Atoi(resourceIdStr)
		if err != nil {
			continue
		}
		resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
		if resourceRecord == nil {
			continue
		}
		resp = append(resp, resourceRecord.(*records.Resource).QiniuUrl)
	}
	httplib.Success(c, map[string]interface{}{"list":resp})
	return
}

type GetCustomerOrdersResponseItem struct {
	ProductId			int		`json:"productId"`
	Payed				int		`json:"payed"`
	Price				int		`json:"price"`
	FirstTitleImage		string		`json:"firstTitleImage"`
	OutTradeNo			string		`json:"outTradeNo"`
}
func customerGetOrders(c *gin.Context) {
	var customerId int
	var ok bool
	if customerId, ok = methods.ParseHttpContextToken(c, consts.Customer); !ok {
		return
	}
	resp := make([]*GetCustomerOrdersResponseItem, 0)
	db := mysql.GetInstance(false)
	orderRecordList := db.Find(records.RecordNameNormalOrder).Select("*").Where("customer_id", "=", customerId).Execute().FetchAll()
	if orderRecordList == nil || orderRecordList.Len() <= 0 {
		httplib.Success(c, map[string]interface{}{"list":resp})
		return
	}
	for _, orderRecord := range orderRecordList.AllRecord() {
		order := orderRecord.(*records.NormalOrder)
		item := new(GetCustomerOrdersResponseItem)
		item.ProductId = order.ProductId
		item.Payed = order.Payed
		item.OutTradeNo = gen32TradeNo(strconv.Itoa(order.ProductOrderId))
		productRecord := db.FindOneByPrimary(records.RecordNameProduct, item.ProductId)
		if productRecord != nil {
			product := productRecord.(*records.Product)
			item.Price = product.Price
			titleResourceIdStrs := strings.Split(product.TitleResourceIds, ",")
			if len(titleResourceIdStrs) > 0 {
				firstResourceId, err := strconv.Atoi(titleResourceIdStrs[0])
				if err == nil {
					resourceRecord := db.FindOneByPrimary(records.RecordNameResource, firstResourceId)
					if resourceRecord != nil {
						item.FirstTitleImage = resourceRecord.(*records.Resource).QiniuUrl
					}
				}
			}

		}
		resp = append(resp, item)
	}
	httplib.Success(c, map[string]interface{}{"list":resp})
	return
}