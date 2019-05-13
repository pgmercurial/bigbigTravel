package records
import (
	"bigbigTravel/component/mysql"
)

const RecordNameProduct = "product"

func init()  {
	var r = &Product{}
	mysql.RegisterRecord(r.Name(), r)
}

type Product struct{
	ProductId	int	`modify:"false" json:"productId" form:"productId" column:"product_id" primary:"true"`
	ProductName	string	`json:"name" form:"name" column:"name" modify:"true"`
	Type	int	`form:"type" column:"type" modify:"true" json:"type"`
	Destination	string	`json:"destination" form:"destination" column:"destination" modify:"true"`
	Count	int	`json:"count" form:"count" column:"count" modify:"true"`
	Price	int	`json:"price" form:"price" column:"price" modify:"true"`
	ValidStartDate	string	`json:"validStartDate" form:"validStartDate" column:"valid_start_date" modify:"true"`
	ValidEndDate	string	`form:"validEndDate" column:"valid_end_date" modify:"true" json:"validEndDate"`
	Show	int	`column:"show" modify:"true" json:"show" form:"show"`
	TitleImageUrl	string	`modify:"true" json:"titleImageUrl" form:"titleImageUrl" column:"titleImageUrl"`
	DetailImageUrl	string	`json:"detailImageUrl" form:"detailImageUrl" column:"detailImageUrl" modify:"true"`
	Remarks	string	`column:"remarks" modify:"true" json:"remarks" form:"remarks"`
	MainTags	string	`json:"mainTags" form:"mainTags" column:"main_tags" modify:"true"`
	SubTags	string	`column:"sub_tags" modify:"true" json:"subTags" form:"subTags"`
	CreateTime	string	`column:"create_time" modify:"false" json:"createTime" form:"createTime"`
	UpdateTime	string	`json:"updateTime" form:"updateTime" column:"update_time" modify:"false"`
}

func (r *Product) Name() string {
	return RecordNameProduct
}
