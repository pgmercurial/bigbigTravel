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
	HeadImages	string	`json:"headImages" form:"headImages" column:"head_images" modify:"true"`
	IntroImages	string	`column:"intro_images" modify:"true" json:"introImages" form:"introImages"`
	Enable	int	`modify:"true" json:"enable" form:"enable" column:"enable"`
	CreateTime	string	`column:"create_time" modify:"false" json:"createTime" form:"createTime"`
	UpdateTime	string	`column:"update_time" modify:"false" json:"updateTime" form:"updateTime"`
}

func (r *SysConf) Name() string {
	return RecordNameSysConf
}
