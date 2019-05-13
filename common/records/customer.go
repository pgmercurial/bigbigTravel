package records
import (
	"bigbigTravel/component/mysql"
)

const RecordNameCustomer = "customer"

func init()  {
	var r = &Customer{}
	mysql.RegisterRecord(r.Name(), r)
}

type Customer struct{
	CustomerId	int	`column:"customer_id" primary:"true" modify:"false" json:"customerId" form:"customerId"`
	OpenId		string	`json:"openId" form:"openId" column:"open_id" modify:"true"`
	CustomerName	string	`json:"name" form:"name" column:"name" modify:"true"`
	Password	string	`json:"password" form:"password" column:"password" modify:"true"`
	HeadPhoto	string	`json:"headPhoto" form:"headPhoto" column:"head_photo" modify:"true"`
	Mobile	string	`json:"mobile" form:"mobile" column:"mobile" modify:"true"`
	Abandon	int	`column:"abandon" modify:"true" json:"abandon" form:"abandon"`
	CreateTime	string	`modify:"false" json:"createTime" form:"createTime" column:"create_time"`
	UpdateTime	string	`json:"updateTime" form:"updateTime" column:"update_time" modify:"false"`
}

func (r *Customer) Name() string {
	return RecordNameCustomer
}
