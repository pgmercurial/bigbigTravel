package records
import (
	"bigbigTravel/component/mysql"
)

const RecordNameResource = "resource"

func init()  {
	var r = &Resource{}
	mysql.RegisterRecord(r.Name(), r)
}

type Resource struct{
	ResourceId	int	`modify:"false" json:"resourceId" form:"resourceId" column:"resource_id" primary:"true"`
	QiniuUrl	string	`json:"qiniuUrl" form:"qiniuUrl" column:"qiniu_url" modify:"true"`
	CreateTime	string	`column:"create_time" modify:"false" json:"createTime" form:"createTime"`
	UpdateTime	string	`form:"updateTime" column:"update_time" modify:"false" json:"updateTime"`
}

func (r *Resource) Name() string {
	return RecordNameResource
}
