package records
import (
	"bigbigTravel/component/mysql"
)

const RecordNameAdmin = "admin"

func init()  {
	var r = &Admin{}
	mysql.RegisterRecord(r.Name(), r)
}

type Admin struct{
	CmsUserId	int	`form:"cmsUserId" column:"cms_user_id" primary:"true" modify:"false" json:"cmsUserId"`
	Role	int	`json:"role" form:"role" column:"role" modify:"true"`
	AdminName	string	`json:"name" form:"name" column:"name" modify:"true"`
	Password	string	`json:"password" form:"password" column:"password" modify:"true"`
	HeadPhoto	string	`json:"headPhoto" form:"headPhoto" column:"head_photo" modify:"true"`
	Mobile	string	`json:"mobile" form:"mobile" column:"mobile" modify:"true"`
	Abandon	int	`json:"abandon" form:"abandon" column:"abandon" modify:"true"`
	CreateTime	string	`column:"create_time" modify:"false" json:"createTime" form:"createTime"`
	UpdateTime	string	`form:"updateTime" column:"update_time" modify:"false" json:"updateTime"`
}

func (r *Admin) Name() string {
	return RecordNameAdmin
}
