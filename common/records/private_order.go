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
	PrivateOrderId	int	`json:"privateOrderId" form:"privateOrderId" column:"private_order_id" primary:"true" modify:"false"`
	CustomerId	int	`form:"customerId" column:"customer_id" modify:"true" json:"customerId"`
	Destination	string	`form:"destination" column:"destination" modify:"true" json:"destination"`
	Withdraw	int	`form:"withdraw" column:"withdraw" modify:"true" json:"withdraw"`
	Handled	int	`json:"handled" form:"handled" column:"handled" modify:"true"`
	CreateTime	string	`json:"createTime" form:"createTime" column:"create_time" modify:"false"`
	UpdateTime	string	`form:"updateTime" column:"update_time" modify:"false" json:"updateTime"`
}

func (r *PrivateOrder) Name() string {
	return RecordNamePrivateOrder
}
