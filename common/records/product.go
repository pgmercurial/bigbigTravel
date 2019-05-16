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
	ProductId	int	`column:"product_id" primary:"true" modify:"false" json:"productId" form:"productId"`
	ProductName	string	`json:"name" form:"name" column:"name" modify:"true"`
	Type	int	`column:"type" modify:"true" json:"type" form:"type"`
	Destination	string	`json:"destination" form:"destination" column:"destination" modify:"true"`
	Count	int	`json:"count" form:"count" column:"count" modify:"true"`
	Price	int	`json:"price" form:"price" column:"price" modify:"true"`
	ValidStartDate	string	`json:"validStartDate" form:"validStartDate" column:"valid_start_date" modify:"true"`
	ValidEndDate	string	`json:"validEndDate" form:"validEndDate" column:"valid_end_date" modify:"true"`
	Show	int	`json:"show" form:"show" column:"show" modify:"true"`
	TitleResourceIds	string	`column:"titleResourceIds" modify:"true" json:"titleResourceIds" form:"titleResourceIds"`
	DetailResourceIds	string	`json:"detailResourceIds" form:"detailResourceIds" column:"detailResourceIds" modify:"true"`
	Remarks	string	`form:"remarks" column:"remarks" modify:"true" json:"remarks"`
	MainTags	string	`modify:"true" json:"mainTags" form:"mainTags" column:"main_tags"`
	SubTags	string	`json:"subTags" form:"subTags" column:"sub_tags" modify:"true"`
	CreateTime	string	`json:"createTime" form:"createTime" column:"create_time" modify:"false"`
	UpdateTime	string	`json:"updateTime" form:"updateTime" column:"update_time" modify:"false"`
}

func (r *Product) Name() string {
	return RecordNameProduct
}
