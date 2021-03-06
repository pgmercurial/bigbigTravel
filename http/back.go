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
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strconv"
	"strings"
)

func init() {
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/login", adminLogin)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/register", adminRegister)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/order/normalList", orderNormalList)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/order/privateList", orderPrivateList)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/order/handlePrivate", orderHandlePrivate)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/product/create", productCreate)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/product/list", productList)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/product/detail", productDetail)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/product/update", productUpdate)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/product/delete", productDelete)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/resource/upload/image", resourceUpload)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/sys-conf/mainTags", sysConfMainTags)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/sys-conf/headImages", sysConfHeadImages)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/sys-conf/introImages", sysConfIntroImages)
	http_middleware.RegisterHttpAction(http_middleware.MethodAll, "admin/mainTagList", adminGetMainTagList)
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
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	req := new(OrderNormalListRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	resp := make([]*OrderListResponseItem, 0)
	//normalOrderRecordList := db.Find(records.RecordNameNormalOrder).Select("*").
	//	OrderBy("product_order_id asc").Limit(20).Offset(req.Page*20).Execute().FetchAll()
	normalOrderRecordList := db.Find(records.RecordNameNormalOrder).Select("*").
		OrderBy("product_order_id asc").Execute().FetchAll()
	totalCount := db.Find(records.RecordNameNormalOrder).Select("*").Count()
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
		item.Payed = nOrder.Payed
		resp = append(resp, item)
	}
	httplib.Success(c, map[string]interface{}{"list":resp, "totalCount":totalCount})
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
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	req := new(OrderPrivateListRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	resp := make([]*OrderPrivateListResponseItem, 0)
	//privateOrderRecordList := db.Find(records.RecordNamePrivateOrder).Select("*").
	//	OrderBy("private_order_id asc").Limit(20).Offset(req.Page*20).Execute().FetchAll()
	privateOrderRecordList := db.Find(records.RecordNamePrivateOrder).Select("*").
		OrderBy("private_order_id asc").Execute().FetchAll()
	if privateOrderRecordList == nil  || privateOrderRecordList.Len() <= 0 {
		httplib.Success(c, map[string]interface{}{"list":resp})
		return
	}
	totalCount := db.Find(records.RecordNamePrivateOrder).Select("*").Count()
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
	httplib.Success(c, map[string]interface{}{"list":resp, "totalCount":totalCount})
	return
}

type OrderHandlePrivateRequest struct {
	PrivateOrderId      	int 		`json:"privateOrderId" form:"privateOrderId"`
	Handled      			int 		`json:"handled" form:"handled"`
}
func orderHandlePrivate(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
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
	TitleImageResourceIds 		string		`json:"titleImageResourceIds" form:"titleImageResourceIds"`
	DetailImageResourceIds 		string		`json:"detailImageResourceIds" form:"detailImageResourceIds"`
	Remarks 			string		`json:"remarks" form:"remarks"`
	MainTags 			string		`json:"mainTags" form:"mainTags"`
	SubTags 			string		`json:"subTags" form:"subTags"`
}
func productCreate(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
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
		TitleResourceIds: req.TitleImageResourceIds,
		DetailResourceIds: req.DetailImageResourceIds,
		Remarks: req.Remarks,
		MainTags: req.MainTags,
		SubTags: req.SubTags,
	}
	db.SaveRecord(product)
	httplib.Success(c)
	return
}

type SysConfMainTagsRequest struct {
	MainTags      	string 		`json:"mainTags" form:"mainTags"`
}
func sysConfMainTags(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	req := new(SysConfMainTagsRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)

	var enabledId int

	sysConfRecord := db.Find(records.RecordNameSysConf).Select("*").Where("enable", "=", 1).Execute().Fetch()
	if sysConfRecord != nil {
		//有则更新
		sysConf := sysConfRecord.(*records.SysConf)
		if req.MainTags != "" {
			sysConf.MainTags = req.MainTags
		}
		db.SaveRecord(sysConf)
		enabledId = sysConf.SysConfId
	} else {
		//无则插入
		enabledId = db.Insert(records.RecordNameSysConf).Columns("main_tags", "enable").Value(req.MainTags, 1).Execute().LastInsertId()
	}

	db.Update(records.RecordNameSysConf).Set("enable", 0).Where("sys_conf_id", "<>", enabledId).Execute()

	httplib.Success(c)
	return
}

type SysConfHeadImagesRequest struct {
	Images   		string 		`json:"images" form:"images"`
}
func sysConfHeadImages(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	req := new(SysConfHeadImagesRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)

	var enabledId int

	sysConfRecord := db.Find(records.RecordNameSysConf).Select("*").Where("enable", "=", 1).Execute().Fetch()
	if sysConfRecord != nil {
		sysConf := sysConfRecord.(*records.SysConf)
		if req.Images != "" {
			sysConf.HeadImages = req.Images
		}
		db.SaveRecord(sysConf)
		enabledId = sysConf.SysConfId
	} else {
		enabledId = db.Insert(records.RecordNameSysConf).Columns("head_images", "enable").Value(req.Images, 1).Execute().LastInsertId()
	}

	db.Update(records.RecordNameSysConf).Set("enable", 0).Where("sys_conf_id", "<>", enabledId).Execute()

	httplib.Success(c)
	return
}

type SysConfIntroImagesRequest struct {
	Images   		string 		`json:"images" form:"images"`
}
func sysConfIntroImages(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	req := new(SysConfIntroImagesRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)

	var enabledId int

	sysConfRecord := db.Find(records.RecordNameSysConf).Select("*").Where("enable", "=", 1).Execute().Fetch()
	if sysConfRecord != nil {
		sysConf := sysConfRecord.(*records.SysConf)
		if req.Images != "" {
			sysConf.IntroImages = req.Images
		}
		db.SaveRecord(sysConf)
		enabledId = sysConf.SysConfId
	} else {
		enabledId = db.Insert(records.RecordNameSysConf).Columns("intro_images", "enable").Value(req.Images, 1).Execute().LastInsertId()
	}

	db.Update(records.RecordNameSysConf).Set("enable", 0).Where("sys_conf_id", "<>", enabledId).Execute()

	httplib.Success(c)
	return
}

func resourceUpload(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	f, fh, err := c.Request.FormFile("file")
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
		fmt.Println(err.Error())
		httplib.Failure(c, exception.ExceptionResourceUploadError, err.Error())
		return
	}

	url := "http://" + conf.Config.Qiniu.Host + "/" + qnResp.Key

	db := mysql.GetInstance(false)
	resourceId := db.Insert(records.RecordNameResource).Columns("qiniu_url").Value(url).Execute().LastInsertId()
	if resourceId <= 0 {
		httplib.Failure(c, exception.ExceptionDBError)
		return
	}

	httplib.Success(c, map[string]interface{}{"resourceId":resourceId})
	return
}


type ProductListRequest struct {
	Page 				int 		`json:"page" form:"page"`

}
type ProductListResponseItem struct {
	ProductId      		int 		`json:"productId" form:"productId"`
	ProductName      	string 		`json:"productName" form:"productName"`
	Type 				int			`json:"type" form:"type"`
	Destination 		string		`json:"destination" form:"destination"`
	Count 				int			`json:"count" form:"count"`
	Price 				int			`json:"price" form:"price"`
	Start 				string		`json:"start" form:"start"`
	End 				string		`json:"end" form:"end"`
	Show 				int			`json:"show" form:"show"`
	TitleImages 		[]*ImageItem		`json:"titleImages" form:"titleImages"`
	DetailImages 		[]*ImageItem		`json:"detailImages" form:"detailImages"`
	Remarks 			string		`json:"remarks" form:"remarks"`
	MainTags 			string		`json:"mainTags" form:"mainTags"`
	SubTags 			string		`json:"subTags" form:"subTags"`
}
type ImageItem struct {
	ResourceId 			int			`json:"resourceId" form:"resourceId"`
	Url 				string		`json:"url" form:"url"`

}
func productList(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	req := new(ProductListRequest)
	resp := make([]*ProductListResponseItem, 0)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	productRecordList := db.Find(records.RecordNameProduct).Select("*").
		Limit(20).Offset(req.Page*20).Execute().FetchAll()
	totalCount := db.Find(records.RecordNameProduct).Select("*").Count()
	if productRecordList == nil {
		httplib.Success(c)
	} else {
		for _, productRecord := range productRecordList.AllRecord() {
			product := productRecord.(*records.Product)
			item := new(ProductListResponseItem)
			item.ProductId = product.ProductId
			item.ProductName = product.ProductName
			item.Type = product.Type
			item.Destination = product.Destination
			item.Count = product.Count
			item.Price = product.Price
			item.Start = product.ValidStartDate
			item.End = product.ValidEndDate
			item.Show = product.Show

			item.TitleImages = make([]*ImageItem, 0)
			titleResourceIdList := strings.Split(product.TitleResourceIds, ",")
			for _, resourceIdStr := range titleResourceIdList {
				resourceId, err := strconv.Atoi(resourceIdStr)
				if err == nil {
					resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
					if resourceRecord != nil {
						newImageItem := new(ImageItem)
						newImageItem.ResourceId = resourceId
						newImageItem.Url = resourceRecord.(*records.Resource).QiniuUrl
						item.TitleImages = append(item.TitleImages, newImageItem)
					}
				}
			}

			item.DetailImages = make([]*ImageItem, 0)
			detailResourceIdList := strings.Split(product.DetailResourceIds, ",")
			for _, resourceIdStr := range detailResourceIdList {
				resourceId, err := strconv.Atoi(resourceIdStr)
				if err == nil {
					resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
					if resourceRecord != nil {
						newImageItem := new(ImageItem)
						newImageItem.ResourceId = resourceId
						newImageItem.Url = resourceRecord.(*records.Resource).QiniuUrl
						item.DetailImages = append(item.DetailImages, newImageItem)
					}
				}
			}

			item.Remarks = product.Remarks
			item.MainTags = product.MainTags
			item.SubTags = product.SubTags
			resp = append(resp, item)
		}
		httplib.Success(c, map[string]interface{}{"list":resp, "totalCount": totalCount})
	}
	return
}


type ProductUpdateRequest struct {
	ProductId      		int 		`json:"productId" form:"productId"`
	ProductName      	string 		`json:"productName" form:"productName"`
	Type 				int			`json:"type" form:"type"`
	Destination 		string		`json:"destination" form:"destination"`
	Count 				int			`json:"count" form:"count"`
	Price 				int			`json:"price" form:"price"`
	Start 				string		`json:"start" form:"start"`
	End 				string		`json:"end" form:"end"`
	Show 				int			`json:"show" form:"show"`
	TitleImageResourceIds 		string		`json:"titleImageResourceIds" form:"titleImageResourceIds"`
	DetailImageResourceIds 		string		`json:"detailImageResourceIds" form:"detailImageResourceIds"`
	Remarks 			string		`json:"remarks" form:"remarks"`
	MainTags 			string		`json:"mainTags" form:"mainTags"`
	SubTags 			string		`json:"subTags" form:"subTags"`
}
func productUpdate(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	req := new(ProductUpdateRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)

	product := new(records.Product)
	productRecord := db.FindOneByPrimary(records.RecordNameProduct, req.ProductId)
	if productRecord != nil {
		product = productRecord.(*records.Product)
	}
	product.ProductName = req.ProductName
	product.Type = req.Type
	product.Destination = req.Destination
	product.Count = req.Count
	product.Price = req.Price
	product.ValidStartDate = req.Start
	product.ValidEndDate = req.End
	product.Show = req.Show
	product.TitleResourceIds = req.TitleImageResourceIds
	product.DetailResourceIds = req.DetailImageResourceIds
	product.Remarks = req.Remarks
	product.MainTags = req.MainTags
	product.SubTags = req.SubTags
	db.SaveRecord(product)
	httplib.Success(c)
	return
}


type ProductDeleteRequest struct {
	ProductId      		int 		`json:"productId" form:"productId"`
}
func productDelete(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	req := new(ProductDeleteRequest)
	httplib.Load(c, req)
	db := mysql.GetInstance(false)
	db.Delete(records.RecordNameProduct).Where("product_id", "=", req.ProductId).Execute()
	httplib.Success(c)
	return
}


type ProductDetailRequest struct {
	ProductId		int			`json:"productId" form:"productId"`
}
type ProductDetailResponse struct {
	ProductName			string	`json:"productName" form:"productName"`
	Type				int		`json:"type" form:"type"`
	Destination			string	`json:"destination" form:"destination"`
	Count				int		`json:"count" form:"count"`
	Price				int		`json:"price" form:"price"`
	ValidStartDate		string	`json:"start" form:"start"`
	ValidEndDate		string	`json:"end" form:"end"`
	Show				int		`json:"show" form:"show"`
	TitleImages			[]*ImageItem	`json:"titleImages" form:"titleImages"`
	DetailImages		[]*ImageItem	`json:"detailImages" form:"detailImages"`
	Remarks				string		`form:"remarks" json:"remarks"`
	MainTags			[]string	`json:"mainTags" form:"mainTags"`
	SubTags				[]string	`json:"subTags" form:"subTags"`
}

func productDetail(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
		return
	}
	req := new(ProductDetailRequest)
	resp := new(ProductDetailResponse)
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
	resp.Destination = product.Destination
	resp.Count = product.Count
	resp.Price = product.Price
	resp.ValidStartDate = product.ValidStartDate
	resp.ValidEndDate = product.ValidEndDate
	resp.Show = product.Show
	resp.Remarks = product.Remarks
	resp.MainTags = strings.Split(product.MainTags, ",")
	resp.SubTags = strings.Split(product.SubTags, ",")

	resp.TitleImages = make([]*ImageItem, 0)
	titleImageIds := strings.Split(product.TitleResourceIds, ",")
	for _, resourceIdStr := range titleImageIds {
		resourceId, _ := strconv.Atoi(resourceIdStr)
		resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
		if resourceRecord != nil {
			item := new(ImageItem)
			item.ResourceId = resourceId
			item.Url = resourceRecord.(*records.Resource).QiniuUrl
			resp.TitleImages = append(resp.TitleImages, item)
		}
	}

	resp.DetailImages = make([]*ImageItem, 0)
	detailImageIds := strings.Split(product.DetailResourceIds, ",")
	for _, resourceIdStr := range detailImageIds {
		resourceId, _ := strconv.Atoi(resourceIdStr)
		resourceRecord := db.FindOneByPrimary(records.RecordNameResource, resourceId)
		if resourceRecord != nil {
			item := new(ImageItem)
			item.ResourceId = resourceId
			item.Url = resourceRecord.(*records.Resource).QiniuUrl
			resp.DetailImages = append(resp.DetailImages, item)
		}
	}

	httplib.Success(c, map[string]interface{}{"detail":resp})
	return
}


func adminGetMainTagList(c *gin.Context) {
	if _, success := methods.ParseHttpContextToken(c, consts.Admin); !success {
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

