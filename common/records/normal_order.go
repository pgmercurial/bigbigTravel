package records
import (
	"bigbigTravel/component/mysql"
)

const RecordNameNormalOrder = "normal_order"

func init()  {
	var r = &NormalOrder{}
	mysql.RegisterRecord(r.Name(), r)
}

type NormalOrder struct{
	ProductOrderId	int	`json:"productOrderId" form:"productOrderId" column:"product_order_id" primary:"true" modify:"false"`
	CustomerId	int	`form:"customerId" column:"customer_id" modify:"true" json:"customerId"`
	ProductId	int	`json:"productId" form:"productId" column:"product_id" modify:"true"`
	Valid	int	`column:"valid" modify:"true" json:"valid" form:"valid"`
	Withdraw	int	`column:"withdraw" modify:"true" json:"withdraw" form:"withdraw"`
	CreateTime	string	`json:"createTime" form:"createTime" column:"create_time" modify:"false"`
	UpdateTime	string	`json:"updateTime" form:"updateTime" column:"update_time" modify:"false"`
}

func (r *NormalOrder) Name() string {
	return RecordNameNormalOrder
}
