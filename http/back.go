package http

import (
	"bigbigTravel/common/methods"
	"bigbigTravel/common/records"
	"bigbigTravel/component/exception"
	"bigbigTravel/component/http/http_middleware"
	"bigbigTravel/component/http/httplib"
	"bigbigTravel/component/mysql"
	"bigbigTravel/component/qiniu"
	"bigbigTravel/conf"
	"bigbigTravel/consts"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func init() {
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/login", adminLogin)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/register", adminRegister)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/order/normalList", orderNormalList)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/order/privateList", orderPrivateList)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/order/handlePrivate", orderHandlePrivate)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/product/create", productCreate)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/product/list", productList)
	//http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/product/update", productUpdate)
	//http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/product/delete", productDelete)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/resource/upload/image", resourceUpload)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/sys-conf", sysConf)
}

type AdminLoginRequest struct {
	UserName      string `json:"userName" form:"userName"`
	Password      string `json:"password" form:"password"`
}
func adminLogin(c *gin.Context) {
	req := new(AdminLoginRequest)
	httplib.Load(c, req)

	db := mysql.GetInstance(false)
	adminRecord := db.Find(records.RecordNameAdmin).Select("*").
		Where("name", "=", req.UserName).
		Where("abandon", "=", 0).Execute().Fetch()
	if adminRecord == nil {
		httplib.Failure(c, exception.ExceptionMissAdmin)
		return
	}
	admin := adminRecord.(*records.Admin)

	if err := methods.VerifyPassword(req.Password, admin.Password); err != nil {
		httplib.Failure(c, exception.ExceptionMissAdmin, err.Error())
		return
	}

	token, _ := methods.GenUserToken(admin.CmsUserId, consts.Admin)
	httplib.Success(c, map[string]interface{}{"token":token})
	return
}

type AdminRegisterRequest struct {
	UserName      string `json:"userName" form:"userName"`
	Password      string `json:"password" form:"password"`
}
func adminRegister(c *gin.Context) {
	req := new(AdminLoginRequest)
	httplib.Load(c, req)

	md5Pwd := methods.Md5Password(req.Password)

	db := mysql.GetInstance(false)
	db.Insert(records.RecordNameAdmin).Columns("name", "password", "abandon").
		Value(req.UserName, md5Pwd, 0).Execute()
	httplib.Success(c)
	return
}

type OrderNormalListRequest struct {
	Page      	int 		`json:"page" form:"page"`
}
type OrderListResponseItem struct {
	OrderId      	int 		`json:"orderId" form:"orderId"`

	Date      		string 		`json:"date" form:"date"`
	CustomerName 	string 		`json:"customerName" form:"customerName"`
	Mobile 			string 		`json:"mobile" form:"mobile"`

	ProductName 	string 		`json:"productName" form:"productName"`
	Payed 			int			`json:"payed" form:"payed"`
}
func orderNormalList(c *gin.Context) {
	req := new(OrderNormalListRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	resp := make([]*OrderListResponseItem, 0)
	normalOrderRecordList := db.Find(records.RecordNameNormalOrder).Select("*").
		OrderBy("product_order_id asc").Limit(20).Offset(req.Page*20).Execute().FetchAll()
	if normalOrderRecordList == nil  || normalOrderRecordList.Len() <= 0 {
		httplib.Success(c, map[string]interface{}{"list":resp})
		return
	}
	for _, nOrderRecord := range normalOrderRecordList.AllRecord() {
		nOrder := nOrderRecord.(*records.NormalOrder)
		item := new(OrderListResponseItem)
		item.OrderId = nOrder.ProductOrderId
		item.Date = nOrder.CreateTime
		customerRecord := db.FindOneByPrimary(records.RecordNameCustomer, nOrder.CustomerId)
		if customerRecord != nil {
			item.CustomerName = customerRecord.(*records.Customer).CustomerName
			item.Mobile = customerRecord.(*records.Customer).Mobile
		}
		productRecord := db.FindOneByPrimary(records.RecordNameProduct, item.OrderId)
		if productRecord != nil {
			item.ProductName = productRecord.(*records.Product).ProductName
		}
		item.Payed = nOrder.Valid
		resp = append(resp, item)
	}
	httplib.Success(c, map[string]interface{}{"list":resp})
	return
}


type OrderPrivateListRequest struct {
	Page      	int 		`json:"page" form:"page"`
}
type OrderPrivateListResponseItem struct {
	OrderId      	int 		`json:"orderId" form:"orderId"`

	Date      		string 		`json:"date" form:"date"`
	CustomerName 	string 		`json:"customerName" form:"customerName"`
	Mobile 			string 		`json:"mobile" form:"mobile"`

	Destination 	string 		`json:"destination" form:"destination"`
	Handled 		int 		`json:"handled" form:"handled"`
}
func orderPrivateList(c *gin.Context) {
	req := new(OrderPrivateListRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	resp := make([]*OrderPrivateListResponseItem, 0)
	privateOrderRecordList := db.Find(records.RecordNamePrivateOrder).Select("*").
		OrderBy("private_order_id asc").Limit(20).Offset(req.Page*20).Execute().FetchAll()
	if privateOrderRecordList == nil  || privateOrderRecordList.Len() <= 0 {
		httplib.Success(c, map[string]interface{}{"list":resp})
		return
	}
	for _, pOrderRecord := range privateOrderRecordList.AllRecord() {
		pOrder := pOrderRecord.(*records.PrivateOrder)
		item := new(OrderPrivateListResponseItem)
		item.OrderId = pOrder.PrivateOrderId
		item.Date = pOrder.CreateTime
		customerRecord := db.FindOneByPrimary(records.RecordNameCustomer, pOrder.CustomerId)
		if customerRecord != nil {
			item.CustomerName = customerRecord.(*records.Customer).CustomerName
			item.Mobile = customerRecord.(*records.Customer).Mobile
		}
		item.Destination = pOrder.Destination
		item.Handled = pOrder.Handled
		resp = append(resp, item)
	}
	httplib.Success(c, map[string]interface{}{"list":resp})
	return
}

type OrderHandlePrivateRequest struct {
	PrivateOrderId      	int 		`json:"privateOrderId" form:"privateOrderId"`
	Handled      			int 		`json:"handled" form:"handled"`
}
func orderHandlePrivate(c *gin.Context) {
	req := new(OrderHandlePrivateRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	db.Update(records.RecordNamePrivateOrder).Set("handled", req.Handled).Where("private_order_id", "=", req.PrivateOrderId).Execute()
	httplib.Success(c)
	return
}

type ProductCreateRequest struct {
	ProductName      	string 		`json:"productName" form:"productName"`
	Type 				int			`json:"type" form:"type"`
	Destination 		string		`json:"destination" form:"destination"`
	Count 				int			`json:"count" form:"count"`
	Price 				int			`json:"price" form:"price"`
	Start 				string		`json:"start" form:"start"`
	End 				string		`json:"end" form:"end"`
	Show 				int			`json:"show" form:"show"`
	TitleImageUrl 		string		`json:"titleImageUrl" form:"titleImageUrl"`
	DetailImageUrl 		string		`json:"detailImageUrl" form:"detailImageUrl"`
	Remarks 			string		`json:"remarks" form:"remarks"`
	MainTags 			string		`json:"mainTags" form:"mainTags"`
	SubTags 			string		`json:"subTags" form:"subTags"`
}
func productCreate(c *gin.Context) {
	req := new(ProductCreateRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	product := &records.Product{
		ProductName: req.ProductName,
		Type: req.Type,
		Destination: req.Destination,
		Count: req.Count,
		Price: req.Price,
		ValidStartDate: req.Start,
		ValidEndDate: req.End,
		Show: req.Show,
		TitleImageUrl: req.TitleImageUrl,
		DetailImageUrl: req.DetailImageUrl,
		Remarks: req.Remarks,
		MainTags: req.MainTags,
		SubTags: req.SubTags,
	}
	db.SaveRecord(product)
	httplib.Success(c)
	return
}

type SysConfRequest struct {
	Op 				int 		`json:"op" form:"op"`

	MainTags      	string 		`json:"mainTags" form:"mainTags"`
	IntroImageUrl   string 		`json:"introImageUrl" form:"introImageUrl"`
}
func sysConf(c *gin.Context) {
	req := new(SysConfRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	if req.Op == 0 { //new
		db.Update(records.RecordNameSysConf).Set("enable", 0).Execute()
		db.Insert(records.RecordNameSysConf).Columns("main_tags", "intro_image_url").Value(req.MainTags, req.IntroImageUrl).Execute()
	} else {  //update
		sysConfRecord := db.Find(records.RecordNameSysConf).Where("enable", "=", 1).Execute().Fetch()
		if sysConfRecord != nil {
			sysConf := sysConfRecord.(*records.SysConf)
			if req.MainTags != "" {
				sysConf.MainTags = req.MainTags
			}
			if req.IntroImageUrl != "" {
				sysConf.IntroImageUrl = req.IntroImageUrl
			}
			db.SaveRecord(sysConf)
		}
	}
	httplib.Success(c)
	return
}

func resourceUpload(c *gin.Context) {
	f, fh, err := c.Request.FormFile("image")
	if err != nil {
		httplib.Failure(c, exception.ExceptionInvalidParams, "miss upload image key `image`:"+err.Error())
		return
	}

	fileBody, err := ioutil.ReadAll(f)
	if err != nil {
		httplib.Failure(c, exception.ExceptionInvalidParams, "file read failed")
		return
	}
	//fileMd5 := helper.Md5(string(fileBody))

	qnResp, err := qiniu.UploadFile(fileBody, "image/"+fh.Filename)
	if err != nil {
		httplib.Failure(c, exception.ExceptionResourceUploadError)
		return
	}

	url := "http://" + conf.Config.Qiniu.Host + "/" + qnResp.Key

	httplib.Success(c, map[string]interface{}{"url":url})
	return
}


type ProductListRequest struct {
	Page 				int 		`json:"page" form:"page"`

}
type ProductListResponseItem struct {
	ProductName      	string 		`json:"productName" form:"productName"`
	Type 				int			`json:"type" form:"type"`
	Destination 		string		`json:"destination" form:"destination"`
	Count 				int			`json:"count" form:"count"`
	Price 				int			`json:"price" form:"price"`
	Start 				string		`json:"start" form:"start"`
	End 				string		`json:"end" form:"end"`
	Show 				int			`json:"show" form:"show"`
	TitleImageUrl 		string		`json:"titleImageUrl" form:"titleImageUrl"`
	DetailImageUrl 		string		`json:"detailImageUrl" form:"detailImageUrl"`
	Remarks 			string		`json:"remarks" form:"remarks"`
	MainTags 			string		`json:"mainTags" form:"mainTags"`
	SubTags 			string		`json:"subTags" form:"subTags"`
}
func productList(c *gin.Context) {
	req := new(ProductListRequest)
	resp := make([]*ProductListResponseItem, 0)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	productRecordList := db.Find(records.RecordNameProduct).Select("*").Where("show","=", 1).
		Limit(20).Offset(req.Page*20).Execute().FetchAll()
	if productRecordList == nil {
		httplib.Success(c)
	} else {
		for _, productRecord := range productRecordList.AllRecord() {
			product := productRecord.(*records.Product)
			item := new(ProductListResponseItem)
			item.ProductName = product.ProductName
			item.Type = product.Type
			item.Destination = product.Destination
			item.Count = product.Count
			item.Price = product.Price
			item.Start = product.ValidStartDate
			item.End = product.ValidEndDate
			item.Show = product.Show
			item.TitleImageUrl = product.TitleImageUrl
			item.DetailImageUrl = product.DetailImageUrl
			item.Remarks = product.Remarks
			item.MainTags = product.MainTags
			item.SubTags = product.SubTags
			resp = append(resp, item)
		}
		httplib.Success(c, map[string]interface{}{"list":resp})
	}
	return
}