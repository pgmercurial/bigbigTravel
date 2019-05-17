package records
import (
	"bigbigTravel/component/mysql"
)

const RecordNamePrivateOrder = "private_order"

func init()  {
	var r = &PrivateOrder{}
	mysql.RegisterRecord(r.Name(), r)
}

type PrivateOrder struct{
	PrivateOrderId	int	`form:"privateOrderId" column:"private_order_id" primary:"true" modify:"false" json:"privateOrderId"`
	CustomerId	int	`modify:"true" json:"customerId" form:"customerId" column:"customer_id"`
	Mobile	string	`json:"mobile" form:"mobile" column:"mobile" modify:"true"`
	CustomerName	string	`json:"name" form:"name" column:"name" modify:"true"`
	Destination	string	`form:"destination" column:"destination" modify:"true" json:"destination"`
	Withdraw	int	`json:"withdraw" form:"withdraw" column:"withdraw" modify:"true"`
	Handled	int	`json:"handled" form:"handled" column:"handled" modify:"true"`
	CreateTime	string	`modify:"false" json:"createTime" form:"createTime" column:"create_time"`
	UpdateTime	string	`modify:"false" json:"updateTime" form:"updateTime" column:"update_time"`
}

func (r *PrivateOrder) Name() string {
	return RecordNamePrivateOrder
}
