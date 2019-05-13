package mysql

import (
	"strings"
	"bigbigTravel/component/helper"
	"fmt"
	"os"
)

func init() {
	RegisterRecord("mysql_table_scheme", new(TableScheme))
}

func generateActionCode(dir, table string, prefix string) *CodeText {
	ct := NewCodeText(dir, table+"_action.go")
	sampleTable := strings.Replace(table, prefix+"_", "", -1)
	controller := strings.Replace(sampleTable, "_", "-", -1)
	fields,primaryField := getFields(table)
	actionName := helper.StrUnderline2CamelCase(sampleTable, true)
	structName := helper.StrUnderline2CamelCase(table, true)
	ct.AddLine(
		`package `+ct.PackageName+`

import (
	"backcms/server"
	"github.com/gin-gonic/gin"
	"github.com/golib/httplib"
	"github.com/golib/mysql"
	"common/records"
	"backcms/routers"
)

func init() {
	routers.RegisterActionParams("record/`+controller+`/create", `+getCreateCheckParams(fields)+`)
	routers.RegisterActionParams("record/`+controller+`/update", `+getUpdateCheckParams(fields, primaryField)+`)
	routers.RegisterActionParams("record/`+controller+`/delete", `+getDeleteCheckParams(primaryField)+`)

	server.RegisterActionPath(server.MethodGet, "record/`+controller+`/list", routers.CheckParamsAction, `+actionName+`ListAction)
	server.RegisterActionPath(server.MethodPost, "record/`+controller+`/list", routers.CheckParamsAction, `+actionName+`ListAction)
	server.RegisterActionPath(server.MethodPost, "record/`+controller+`/create", routers.CheckParamsAction, `+actionName+`CreateAction)
	server.RegisterActionPath(server.MethodPost, "record/`+controller+`/update", routers.CheckParamsAction, `+actionName+`UpdateAction)
	server.RegisterActionPath(server.MethodPost, "record/`+controller+`/delete", routers.CheckParamsAction, `+actionName+`DeleteAction)
	server.RegisterActionPath(server.MethodGet, "record/`+controller+`/delete", routers.CheckParamsAction, `+actionName+`DeleteAction)
}

func `+actionName+`ListAction(context *gin.Context)  {
	db := mysql.GetInstance()
	mp := make(map[string]string)
	for k, v := range context.Request.Form {
		mp[k] = v[0]
	}
	recordList := db.SearchRecord(records.RecordName`+structName+`,mp).AllRecord()
	httplib.Success(context, map[string]interface{}{
		"list":recordList,
		})
}

func `+actionName+`CreateAction(context *gin.Context)  {
	record := new(records.`+structName+`)
	httplib.Load(context, record)
	record.`+helper.Ucfirst(primaryField.Key)+` = 0
	db := mysql.GetInstance()
	db.SaveRecord(record)
	httplib.Success(context, record)
}

func `+actionName+`UpdateAction(context *gin.Context)  {
	record := new(records.`+structName+`)
	httplib.Load(context, record)
	db := mysql.GetInstance()
	columns := routers.GetUploadedParams(context)
	i := db.SaveRecord(record, columns...)
	if i == 0 {
		httplib.Success(context)
		return
	}else {
		db.LoadRecord(record)
		httplib.Success(context, record)
	}
}

func `+actionName+`DeleteAction(context *gin.Context)  {
	record := new(records.`+structName+`)
	httplib.Load(context, record)
	db := mysql.GetInstance()
	db.DeleteRecord(record)
	httplib.Success(context)
}
`)
	return ct
}

func getFields(table string) ([]*RecordField, *RecordField) {
	db := GetInstance(false)
	fmt.Println(table)
	rl := db.Query("mysql_table_scheme", "desc `"+table+"`")
	result := make([]*RecordField, 0)
	var primaryField *RecordField
	for i:=0 ; i < rl.Len(); i++ {
		rf := new(RecordField)
		field := rl.GetAt(i).(*TableScheme)
		rf.Key = helper.StrUnderline2CamelCase(field.Field, false)
		rf.Column = field.Field
		rf.IsPrimary = field.Key == "PRI"
		rf.Modify = !(rf.IsPrimary || field.Field == "create_time" || field.Field  =="update_time")
		rf.Index = i
		if rf.IsPrimary {
			primaryField = rf
		}else {
			result = append(result, rf)
		}
	}
	return result,primaryField
}

func getCreateCheckParams(rs []*RecordField) string {
	ps := make([]string, 0, len(rs))
	for _,r := range rs{
		if r.Modify {
			ps = append(ps, r.Key)
		}
	}

	return `"`+strings.Join(ps, `","`)+`"`
}

func getUpdateCheckParams(rs []*RecordField, pri *RecordField) string {
	ps := make([]string, 0, len(rs))
	for _,r := range rs{
		if r.Modify {
			ps = append(ps, r.Key)
		}
	}

	return `"`+pri.Key+`", []string{"`+strings.Join(ps, `","`)+`"}`
}

func getDeleteCheckParams(pri *RecordField) string {

	return `"`+pri.Key+`"`
}

func GenerateSignalAction(dir string, name string, prefix string)  {
	fmt.Println("Generating CURD Action:",name)
	codeText := generateActionCode(dir, name, prefix)
	_,err := os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir, 0755)
		if err != nil{
			fmt.Println(err.Error())
			os.Exit(0)
		}
	}
	filePath := strings.TrimRight(dir, "/")+"/"+codeText.FileName
	f,err := os.Create(filePath)
	checkErr(err)
	_,err = f.WriteString(strings.Join(codeText.Lines,"\n"))
	checkErr(err)
	fmt.Println("Success:", filePath)
}

func GenerateAllAction(dir string, prefix string)  {
	db := GetInstance(false)
	rows, err := db.Db().Query("show tables")
	checkErr(err)
	var tableName string
	for rows.Next() {
		rows.Scan(&tableName)
		GenerateSignalAction(dir, tableName, prefix)
	}
}


