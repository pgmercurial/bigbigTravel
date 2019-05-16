package records
import (
	"bigbigTravel/component/mysql"
)

const RecordNameSysConf = "sys_conf"

func init()  {
	var r = &SysConf{}
	mysql.RegisterRecord(r.Name(), r)
}

type SysConf struct{
	SysConfId	int	`json:"sysConfId" form:"sysConfId" column:"sys_conf_id" primary:"true" modify:"false"`
	MainTags	string	`json:"mainTags" form:"mainTags" column:"main_tags" modify:"true"`
	Enable	int	`form:"enable" column:"enable" modify:"true" json:"enable"`
	CreateTime	string	`json:"createTime" form:"createTime" column:"create_time" modify:"false"`
	UpdateTime	string	`json:"updateTime" form:"updateTime" column:"update_time" modify:"false"`
}

func (r *SysConf) Name() string {
	return RecordNameSysConf
}
