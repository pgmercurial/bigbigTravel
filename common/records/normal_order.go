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
	CustomerId	int	`modify:"true" json:"customerId" form:"customerId" column:"customer_id"`
	Mobile	string	`json:"mobile" form:"mobile" column:"mobile" modify:"true"`
	CustomerName	string	`json:"name" form:"name" column:"name" modify:"true"`
	ProductId	int	`form:"productId" column:"product_id" modify:"true" json:"productId"`
	Withdraw	int	`form:"withdraw" column:"withdraw" modify:"true" json:"withdraw"`
	Payed	int	`json:"payed" form:"payed" column:"payed" modify:"true"`
	CreateTime	string	`modify:"false" json:"createTime" form:"createTime" column:"create_time"`
	UpdateTime	string	`json:"updateTime" form:"updateTime" column:"update_time" modify:"false"`
}

func (r *NormalOrder) Name() string {
	return RecordNameNormalOrder
}
